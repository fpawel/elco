package app

import (
	"context"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/powerman/structlog"
	"time"
)

var (
	log      = structlog.New()
	ctxApp   context.Context
	hardware = Hardware{
		continueFunc:  func() {},
		cancelFunc:    func() {},
		skipDelayFunc: func() {},
		ctx:           context.Background(),
	}
	portMeasurer = comport.NewReadWriter(func() comport.Config {
		return comport.Config{
			Baud:        115200,
			ReadTimeout: time.Millisecond,
			Name:        cfg.Cfg.User().ComportMeasurer,
		}
	}, func() comm.Config {
		return cfg.Cfg.Predefined().ComportMeasurer
	})

	portGas = comport.NewReadWriter(func() comport.Config {
		return comport.Config{
			Baud:        9600,
			ReadTimeout: time.Millisecond,
			Name:        cfg.Cfg.User().ComportGas,
		}
	}, func() comm.Config {
		return cfg.Cfg.Predefined().ComportGas
	})
)
