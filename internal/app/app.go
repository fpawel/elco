package app

import (
	"context"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/ktx500"
	"github.com/fpawel/elco/internal/peer"
	"github.com/fpawel/gohelp/must"
	"github.com/lxn/win"
	_ "github.com/mattn/go-sqlite3"
	"github.com/powerman/structlog"
	"net/rpc"
	"sync"
)

type App struct{}

func Run() error {

	peer.AssertRunOnes()
	data.Open()

	var cancel func()
	ctxApp, cancel = context.WithCancel(context.Background())

	for _, svcObj := range []interface{}{
		new(api.PartiesCatalogueSvc),
		new(api.LastPartySvc),
		new(api.ProductTypesSvc),
		api.NewProductFirmware(runner{}),
		new(api.PdfSvc),
		&api.RunnerSvc{Runner: runner{}},
		api.NewPeerSvc(peerNotifier{}),
		new(api.ConfigSvc),
		new(api.ProductsCatalogueSvc),
	} {
		must.AbortIf(rpc.Register(svcObj))
	}

	closeHttpServer := startHttpServer()

	peer.Init("")

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
	peer.Close()
	data.Close()
	return nil
}

var (
	ctxApp         context.Context
	cancelWorkFunc = func() {}
	skipDelayFunc  = func() {}
	wgWork         sync.WaitGroup
	log            = structlog.New()
)
