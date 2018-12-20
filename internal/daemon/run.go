package daemon

import (
	"context"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/goutils/intrng"
	"github.com/sirupsen/logrus"
	"sync"
)

func (x *D) RunMainError() {
	x.runHardware("Снятие основной погрешности", x.determineMainError)
}

func (x *D) RunTemperature(workCheck [3]bool) {
	x.runHardware("Снятие термокомпенсации", func() error {
		for i, temperature := range []data.Temperature{20, -20, 50} {
			if workCheck[i] {
				notify.Statusf(x.w, "%v⁰C: снятие термокомпенсации", temperature)
				if err := x.determineTemperature(temperature); err != nil {
					logrus.WithField("thermal_compensation_determination", temperature).Errorf("%+v", err)
					return err
				}
			}
		}
		return x.determineNKU2()
	})
}

func (x *D) StopHardware() {
	x.hardware.cancel()
}

func (x *D) SkipDelay() {
	x.hardware.skipDelay()
	logrus.Warn("пользователь перрвал задержку")
}

func (x *D) RunReadCurrent(checkPlaces [12]bool) {
	var places, xs []int
	for i, v := range checkPlaces {
		if v {
			places = append(places, i)
			xs = append(xs, i+1)
		}
	}
	x.runHardware("опрос блоков измерительных: "+intrng.Format(xs), func() error {
		x.port.measurer.SetLog(x.sets.Config().Predefined.Work.CaptureComport)
		for {
			for _, place := range places {
				if _, err := x.readMeasure(place); err != nil {
					return err
				}
			}
		}
	})
}

type WorkFunc = func() error

func (x *D) runHardware(what string, work WorkFunc) {
	x.hardware.cancel()
	x.hardware.WaitGroup.Wait()
	x.hardware.WaitGroup = sync.WaitGroup{}
	x.hardware.ctx, x.hardware.cancel = context.WithCancel(x.ctx)

	cfg := x.sets.Config()

	notify.HardwareStarted(x.w, what)
	x.hardware.WaitGroup.Add(1)

	go func() {

		if err := x.port.measurer.Open(cfg.Comport.Measurer, 115200, 0, x.hardware.ctx); err != nil {
			notify.HardwareErrorf(x.w, "%s: %v", cfg.Comport.Measurer, err)
		} else {
			x.port.measurer.SetLog(true)
			if err := work(); err != nil && x.hardware.ctx.Err() != context.Canceled {
				notify.HardwareErrorf(x.w, "%s: %v", what, err)
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

			if err := x.switchGas(0); err != nil {
				notify.HardwareErrorf(x.w, "отключение газового блока при завершении: %s: %v", x.port.gas.Config().Name, err)
			}

			if err := x.port.gas.Close(); err != nil {
				notify.HardwareErrorf(x.w, "%s: %v", x.port.gas.Config().Name, err)
			}
		}
		notify.HardwareStoppedf(x.w, "выполнение окончено: %s", what)
		x.hardware.WaitGroup.Done()
	}()
}
