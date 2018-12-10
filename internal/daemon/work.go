package daemon

import (
	"context"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/goutils/serial-comm/comport"
	"github.com/fpawel/goutils/serial-comm/modbus"
	"github.com/pkg/errors"
)

func (x *D) Continue() {
	x.hardware.Continue()
}

func (x *D) switchGas(n int) error {
	c := x.sets.Config()
	//if !x.port.gas.Opened() {
	//	err := x.port.gas.Open(c.Comport.GasSwitcher, 9600, 0, x.hardware.ctx)
	//	if err != nil {
	//		return errors.Wrap(err, "не удалось открыть СОМ порт газового блока")
	//	}
	//}
	req := modbus.NewSwitchGasOven(byte(n))
	if _, err := x.port.gas.GetResponse(req.Bytes(), c.GasSwitcher); err != nil {
		return errors.Wrapf(err, "нет связи c газовым блоком через %s", c.Comport.GasSwitcher)
	}
	return nil
}

func (x *D) readMeasure(place int) ([]float64, error) {
	c := x.sets.Config()

	values, err := modbus.Read3BCDValues(comport.Comm{
		Port:   x.port.measurer,
		Config: c.Measurer,
	}, modbus.Addr(place+101), 0, 8)

	switch err {

	case nil:
		notify.ReadCurrent(x.w, api.ReadCurrent{
			Place:  place,
			Values: values,
		})

	case context.Canceled:

	case context.DeadlineExceeded:

		notify.HardwareErrorf(x.w, "блок измерения №%d: не отечает: %s",
			place+1,
			x.port.measurer.Dump())

	default:
		notify.HardwareErrorf(x.w, "блок измерения №%d: %+v: %s",
			place+1, err,
			x.port.measurer.Dump())

	}
	return values, err

}
