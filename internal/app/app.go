package app

import (
	"context"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/ktx500"
	"github.com/fpawel/elco/internal/peer"
	"github.com/lxn/win"
	_ "github.com/mattn/go-sqlite3"
	"sync"
	"time"
)

type App struct{}

func Run() error {

	closeHttpServer := startHttpServer()

	peer.Init("")

	var cancel func()
	ctxApp, cancel = context.WithCancel(context.Background())

	go ktx500.TraceTemperature()

	// цикл оконных сообщений
	for {
		var msg win.MSG
		if win.GetMessage(&msg, 0, 0, 0) == 0 {
			break
		}
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}

	cancel()
	closeHttpServer()

	return nil
}

var (
	ctxApp         context.Context
	ctxWork        context.Context
	cancelWorkFunc = func() {}
	skipDelayFunc  = func() {}
	wgWork         sync.WaitGroup
	portMeasurer   = comport.NewReadWriter(func() comport.Config {
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
