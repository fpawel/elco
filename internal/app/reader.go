package app

import (
	"context"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/powerman/structlog"
)

type reader struct {
	*comport.Reader
	Config comm.Config
	Ctx    context.Context
}

func (x reader) GetResponse(logger *structlog.Logger, request []byte, responseParser comm.ResponseParser) ([]byte, error) {

	return x.Reader.GetResponse(comm.Request{
		Logger:         logger,
		Bytes:          request,
		Config:         x.Config,
		ResponseParser: responseParser,
	}, x.Ctx)
}

func gasBlockReader() modbus.ResponseReader {
	return reader{
		Reader: portGas,
		Config: cfg.Cfg.Predefined().ComportGas,
		Ctx:    hardware.ctx,
	}
}

func measurerReader(ctx context.Context) modbus.ResponseReader {
	return reader{
		Reader: portMeasurer,
		Config: cfg.Cfg.Predefined().ComportMeasurer,
		Ctx:    ctx,
	}
}
