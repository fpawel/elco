package daemon

import (
	"context"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/goutils/intrng"
	"sync"
)

func (x *D) RunMainWork(workCheck [5]bool) {
	x.runHardware(Work{"Настройка ЭХЯ", func() error {
		for i, w := range x.mainWorks() {
			if workCheck[i] {
				if err := w.Func(); err != nil {
					return err
				}
			}
		}
		return nil
	}})
}

func (x *D) StopHardware() {
	x.hardware.cancel()
}

func (x *D) SkipDelay() {
	x.hardware.skipDelay()
}

func (x *D) RunReadCurrent(checkPlaces [12]bool) {
	var places, xs []int
	for i, v := range checkPlaces {
		if v {
			places = append(places, i)
			xs = append(xs, i+1)
		}
	}
	x.runHardware(Work{"опрос блоков измерительных: " + intrng.Format(xs), func() error {
		for {
			for _, place := range places {
				if _, err := x.readMeasure(place); err != nil {
					return err
				}
			}
		}
	}})
}

func (x *D) runHardware(work Work) {
	x.hardware.cancel()
	x.hardware.WaitGroup.Wait()
	x.hardware.WaitGroup = sync.WaitGroup{}
	x.hardware.ctx, x.hardware.cancel = context.WithCancel(x.ctx)

	cfg := x.sets.Config()

	notify.HardwareStarted(x.w, work.Name)
	x.hardware.WaitGroup.Add(1)

	go func() {

		if err := x.port.measurer.Open(cfg.Comport.Measurer, 115200, 0, x.hardware.ctx); err != nil {
			notify.HardwareErrorf(x.w, "%s: %v", cfg.Comport.Measurer, err)
		} else {
			if err := work.Func(); err != nil && x.hardware.ctx.Err() != context.Canceled {
				notify.HardwareErrorf(x.w, "%s: %v", work.Name, err)
			}
		}

		if x.port.measurer.Opened() {
			if err := x.port.measurer.Close(); err != nil {
				notify.HardwareErrorf(x.w, "%s: %v", x.port.measurer.Config().Name, err)
			}
		}
		if x.port.gas.Opened() {
			if err := x.port.gas.Close(); err != nil {
				notify.HardwareErrorf(x.w, "%s: %v", x.port.gas.Config().Name, err)
			}
		}
		notify.HardwareStoppedf(x.w, "выполнение окончено: %s", work.Name)
		x.hardware.WaitGroup.Done()
	}()
}
