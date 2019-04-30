package daemon

import (
	"context"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
)

type responseReader struct {
	Reader *comport.Reader
	Config comm.Config
	Ctx    context.Context
}

func (x responseReader) GetResponse(request []byte, prs comm.ResponseParser) ([]byte, error) {
	return x.Reader.GetResponse(request, x.Config, x.Ctx, prs)
}

func (x *D) gasBlockReader() modbus.ResponseReader {
	return responseReader{
		Reader: x.portGas,
		Config: x.cfg.Predefined().ComportGas,
		Ctx:    x.hardware.ctx,
	}
}

func (x *D) measurerReader(ctx context.Context) modbus.ResponseReader {
	return responseReader{
		Reader: x.portMeasurer,
		Config: x.cfg.Predefined().ComportMeasurer,
		Ctx:    ctx,
	}
}
