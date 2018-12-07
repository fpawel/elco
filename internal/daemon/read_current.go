package daemon

import (
	"context"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/goutils/serial/comport"
	"github.com/fpawel/goutils/serial/modbus"
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

	cfg := x.sets.Config()

	var ctx context.Context
	ctx, x.hardware.cancel = context.WithCancel(x.ctx)
	port := new(comport.Port)

	if err := port.Open(cfg.ComportHardware, ctx); err != nil {
		notify.HardwareErrorf(x.w, "%s: %v", cfg.ComportHardware.Serial.Name, err.Error())
		return
	}

	notify.HardwareStarted(x.w, "опрос стенда")

	go func() {

		defer func() {
			if err := port.Close(); err != nil {
				notify.HardwareErrorf(x.w, "%s: %v", cfg.ComportHardware.Serial.Name, err.Error())
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
					cfg := x.sets.Config()
					if !cfg.BlockSelected[place] {
						continue
					}
					notify.Statusf(x.w, "опрос: блок %d", place+1)
					values, err := modbus.Read3BCDValues(port, modbus.Addr(place+101), 0, 8)
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
							cfg.ComportHardware.Serial.Name, place+1,
							port.Dump())
						return
					default:
						notify.HardwareErrorf(x.w, "%s: стенд 6364, блок %d: %+v: %v",
							cfg.ComportHardware.Serial.Name, place+1, err,
							port.Dump())
						return
					}

				}
			}
		}

	}()
}

func read6364(port *comport.Port, addr modbus.Addr) ([]float64, error) {
	return modbus.Read3BCDValues(port, addr, 0, 8)
}
