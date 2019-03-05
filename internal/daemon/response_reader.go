package daemon

import (
	"context"
	"github.com/fpawel/elco/pkg/serial-comm/comm"
	"github.com/fpawel/elco/pkg/serial-comm/comport"
	"github.com/fpawel/elco/pkg/serial-comm/modbus"
)

type responseReader struct {
	Port   *comport.Port
	Config comm.Config
	Ctx    context.Context
}

func (x responseReader) GetResponse(request []byte, prs comm.ResponseParser) ([]byte, error) {
	return x.Port.GetResponse(request, x.Config, x.Ctx, prs)
}

func (x *D) gasBlockReader() modbus.ResponseReader {
	return responseReader{
		Port:   x.portGas,
		Config: x.cfg.Predefined().ComportGas,
		Ctx:    x.hardware.ctx,
	}
}

func (x *D) measurerReader(ctx context.Context) modbus.ResponseReader {
	return responseReader{
		Port:   x.portMeasurer,
		Config: x.cfg.Predefined().ComportMeasurer,
		Ctx:    ctx,
	}
}
