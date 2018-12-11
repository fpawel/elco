package daemon

import (
	"context"
	"fmt"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/goutils/intrng"
	"github.com/fpawel/goutils/serial-comm/comport"
	"github.com/fpawel/goutils/serial-comm/modbus"
	"github.com/pkg/errors"
	"sync"
)

func (x *D) switchGas(n int) error {
	err := x.doSwitchGas(n)
	if err == nil {
		return nil
	}
	s := "Не удалось "
	if n == 0 {
		s += "отключить газ"
	} else {
		s += fmt.Sprintf("подать ПГС%d", n)
	}
	s += ": " + err.Error() + ".\n\n"

	if n == 0 {
		s += "Отключите газ"
	} else {
		s += fmt.Sprintf("Подайте ПГС%d", n)
	}
	s += " вручную."
	notify.Warning(x.w, s)
	if x.hardware.ctx.Err() == context.Canceled {
		return err
	}
	return nil
}

func (x *D) doSwitchGas(n int) error {
	c := x.sets.Config()
	if !x.port.gas.Opened() {
		if err := x.port.gas.Open(c.Comport.GasSwitcher, 9600, 0, x.hardware.ctx); err != nil {
			return err
		}
	}
	req := modbus.NewSwitchGasOven(byte(n))
	_, err := x.port.gas.GetResponse(req.Bytes(), c.GasSwitcher)
	if err == context.DeadlineExceeded {
		err = errors.New("нет ответа от газового блока: " + x.port.gas.Dump())
	}
	return err
}

func (x *D) readMeasure(place int) ([]float64, error) {

	c := x.sets.Config()

	switch values, err := modbus.Read3BCDValues(comport.Comm{
		Port:   x.port.measurer,
		Config: c.Measurer,
	}, modbus.Addr(place+101), 0, 8); err {

	case nil:

		notify.ReadCurrent(x.w, api.ReadCurrent{
			Place:  place,
			Values: values,
		})
		return values, nil

	case context.Canceled:
		return nil, context.Canceled

	case context.DeadlineExceeded:

		return nil, errors.Errorf("блок измерения №%d: не отечает: %s",
			place+1,
			x.port.measurer.Dump())

	default:
		return nil, errors.Errorf("блок измерения №%d: %+v: %s",
			place+1, err,
			x.port.measurer.Dump())
	}
}

func (x *D) StopHardware() {
	x.hardware.cancel()
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

func (x *D) mainWork(workCheck [5]bool) error {
	for i, work := range [5]Work{
		{
			"20\"C",
			func() error {
				return nil
			},
		},
		{
			"-20\"C",
			func() error {
				return nil
			},
		},
		{
			"+50\"C",
			func() error {
				return nil
			},
		},
		{
			"\"Прошивка\"",
			func() error {
				return nil
			},
		},
		{
			"Проверка",
			func() error {
				return nil
			},
		},
	} {
		if workCheck[i] {
			if err := work.Func(); err != nil {
				return err
			}
		}
	}
}

type Work struct {
	Name string
	Func WorkFunc
}
type WorkFunc = func() error
