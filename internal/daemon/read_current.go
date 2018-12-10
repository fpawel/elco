package daemon

import (
	"context"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/goutils/intrng"
	"sync"
)

func (x *D) StopHardware() {
	x.hardware.cancel()
	x.hardware.WaitGroup.Wait()

}

func (x *D) RunReadCurrent(checkPlaces [12]bool) {

	x.hardware.cancel()
	x.hardware.WaitGroup.Wait()
	x.hardware.WaitGroup = sync.WaitGroup{}
	x.hardware.WaitGroup.Add(1)

	c := x.sets.Config()

	x.hardware.ctx, x.hardware.cancel = context.WithCancel(x.ctx)

	if err := x.port.measurer.Open(c.Comport.Measurer, 115200, 0, x.hardware.ctx); err != nil {
		notify.HardwareErrorf(x.w, "%s: %v", c.Comport.Measurer, err.Error())
		return
	}

	var places []int
	for i, v := range checkPlaces {
		if v {
			places = append(places, i)
		}
	}

	notify.HardwareStarted(x.w, "опрос блоков измерительных: "+intrng.Format(places))

	go func() {
		for {
			for _, place := range places {
				notify.Statusf(x.w, "опрос: блок измерения №%d", place+1)
				if _, err := x.readMeasure(place); err != nil {
					if err := x.port.measurer.Close(); err != nil {
						notify.HardwareErrorf(x.w, "%s: %s", c.Comport.Measurer, err.Error())
					}
					notify.HardwareStoppedf(x.w, "завершён опрос блоков измерительных: %s", intrng.Format(places))
					x.hardware.WaitGroup.Done()
					return
				}
			}
		}
	}()
}
