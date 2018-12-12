package daemon

import (
	"context"
	"fmt"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/goutils/serial-comm/comport"
	"github.com/fpawel/goutils/serial-comm/modbus"
	"github.com/pkg/errors"
	"time"
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

func (x *D) blowGas(nGas int) error {
	if err := x.switchGas(nGas); err != nil {
		return err
	}
	var ctx context.Context
	ctx, x.hardware.skipDelay = context.WithCancel(x.hardware.ctx)

	timeMinutes := x.sets.Config().Work.BlowGasMinutes

	t := time.After(time.Minute * time.Duration(timeMinutes))

	notify.Delay(x.w, api.DelayInfo{
		Run:         true,
		What:        fmt.Sprintf("Продувка ПГС%d", nGas),
		TimeSeconds: timeMinutes * 60,
	})

	defer notify.Delay(x.w, api.DelayInfo{Run: false})

	for {
		select {

		case <-ctx.Done():
			return nil

		case <-t:
			return nil

		default:
			time.Sleep(50 * time.Millisecond)
		}
	}

}

func (x *D) mainWorks() [5]Work {

	return [5]Work{
		{
			"20\"C",
			func() error {
				if err := x.blowGas(1); err != nil {
					return err
				}
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
	}
}

type Work struct {
	Name string
	Func WorkFunc
}
type WorkFunc = func() error
