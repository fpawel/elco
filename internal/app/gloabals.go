package app

import (
	"context"
	"github.com/fpawel/comm/comport"
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
	portMeasurer = comport.NewReader(comport.Config{
		Baud:        115200,
		ReadTimeout: time.Millisecond,
	})

	portGas = comport.NewReader(comport.Config{
		Baud:        9600,
		ReadTimeout: time.Millisecond,
	})
)
