package daemon

import (
	"context"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/goutils/serial-comm/comport"
	"github.com/fpawel/goutils/serial-comm/modbus"
	"sync"
)

func (x *D) StopHardware() {
	x.hardware.cancel()
	x.hardware.WaitGroup.Wait()
	notify.HardwareStopped(x.w, "выполнение прервано пользователем")
}

func (x *D) RunReadCurrent() {

	x.hardware.cancel()
	x.hardware.WaitGroup.Wait()
	x.hardware.WaitGroup = sync.WaitGroup{}
	x.hardware.WaitGroup.Add(1)

	c := x.sets.Config()

	portCfg := comport.SerialConfig(c.ComportName, 115200)

	var ctx context.Context
	ctx, x.hardware.cancel = context.WithCancel(x.ctx)
	port := new(comport.Port)

	if err := port.Open(portCfg, 0, ctx); err != nil {
		notify.HardwareErrorf(x.w, "%s: %v", c.ComportName, err.Error())
		return
	}

	notify.HardwareStarted(x.w, "опрос стенда")

	go func() {

		defer func() {
			if err := port.Close(); err != nil {
				notify.HardwareErrorf(x.w, "%s: %v", c.ComportName, err.Error())
				return
			}
			notify.HardwareStopped(x.w, "опрос стенда завершён")
			x.hardware.WaitGroup.Done()
		}()

		for {
			for place := 0; place < 12; place++ {
				select {
				case <-ctx.Done():
					return
				default:

					notify.Statusf(x.w, "опрос: блок %d", place+1)

					responseReader := comport.Comm{
						Port:   port,
						Config: c.Measurer.Comm,
					}

					values, err := modbus.Read3BCDValues(responseReader, modbus.Addr(place+101), 0, 8)

					switch err {
					case nil:
						notify.ReadCurrent(x.w, api.ReadCurrent{
							Place:  place,
							Values: values,
						})
					case context.Canceled:
						return
					case context.DeadlineExceeded:
						notify.HardwareErrorf(x.w, "%s: стенд 6364: блок %d не отвечает: %s",
							c.ComportName, place+1,
							port.Dump())
						return
					default:
						notify.HardwareErrorf(x.w, "%s: стенд 6364, блок %d: %+v: %v",
							c.ComportName, place+1, err,
							port.Dump())
						return
					}

				}
			}
		}

	}()
}
