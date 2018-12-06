package daemon

import (
	"context"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/goutils/serial/comport"
	"github.com/fpawel/goutils/serial/modbus"
	"sync"
)

func (x *D) RunReadCurrent() {
	x.comports.cancel()
	x.comports.WaitGroup.Wait()
	x.comports.WaitGroup = sync.WaitGroup{}
	x.comports.WaitGroup.Add(1)
	x.comports.Context, x.comports.cancel = context.WithCancel(context.Background())

	go func() {

		defer func() {
			x.comports.cancel()
			x.comports.WaitGroup.Done()
		}()

		cfg := x.sets.Config()
		port, err := x.comports.Open(cfg.ComportHardware, x.comports.Context)
		if err != nil {
			notify.HardwareErrorf(x.w, "%s: %v", cfg.ComportHardware.Serial.Name, err.Error())
			return
		}
		for {
			var v api.ReadCurrent
			for v.Place = 0; v.Place < 12; v.Place++ {
				select {
				case <-x.comports.Context.Done():
					return
				default:
					cfg := x.sets.Config()
					if !cfg.BlockSelected[v.Place] {
						continue
					}
					v.Values, err = modbus.Read3BCDValues(port, modbus.Addr(v.Place+101), 0, 8)
					if err != nil {
						notify.HardwareErrorf(x.w, "%s: %v: %v", cfg.ComportHardware.Serial.Name, err.Error(), port.Dump())
						return
					}
					notify.ReadCurrent(x.w, v)
				}
			}
		}

	}()
}

func read6364(port *comport.Port, addr modbus.Addr) ([]float64, error) {
	return modbus.Read3BCDValues(port, addr, 0, 8)
}
