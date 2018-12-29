package daemon

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/goutils/serial-comm/comport"
	"github.com/fpawel/goutils/serial-comm/modbus"
)

func (x *D) readBlockMeasure(block int) ([]float64, error) {

	c := x.sets.Config()

	values, err := modbus.Read3BCDValues(comport.Comm{
		Port:   x.port.measurer,
		Config: c.Measurer,
	}, modbus.Addr(block+101), 0, 8)

	switch err {

	case nil:
		notify.ReadCurrent(x.w, api.ReadCurrent{
			Block:  block,
			Values: values,
		})
		return values, nil

	case context.Canceled:
		return nil, context.Canceled

	case context.DeadlineExceeded:
		w := x.port.measurer.LastWork()
		return nil, errorMeasurer.Here().Append("не отечает")

	default:

		return nil, errorMeasurer.Here().Append(err.Error())
	}
}
