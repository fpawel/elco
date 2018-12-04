package daemon

import (
	"context"
	"github.com/fpawel/goutils/serial/comport"
	"github.com/fpawel/goutils/serial/modbus"
	"sync"
)

func (x *D) runRead6364() {
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

		}
	}()
}

func read6364(port *comport.Port, addr modbus.Addr) ([]float64, error) {
	return modbus.Read3BCDValues(port, addr, 0, 8)
}
