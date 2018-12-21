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
	w := x.port.measurer.LastWork()

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
		return nil, merry.Errorf("блок измерения %d: не отечает: %s", block+1,
			x.port.measurer.LastWork().
				FormatRequest()).
			WithValue("block", block).
			WithValue("request", fmt.Sprintf("% X", w.Request)).
			WithValue("duration", w.Duration)

	default:

		return nil, merry.Appendf(err, "блок измерения %d: %s", block+1,
			x.port.measurer.LastWork().FormatResponse()).
			WithValue("block", block).
			WithValue("request", fmt.Sprintf("% X", w.Request)).
			WithValue("response", fmt.Sprintf("% X", w.Response)).
			WithValue("duration", w.Duration)
	}
}
