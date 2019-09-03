package app

import (
	"context"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/ktx500"
	"github.com/fpawel/elco/internal/peer"
	"github.com/fpawel/gohelp/must"
	"github.com/fpawel/gohelp/winapp"
	"github.com/lxn/win"
	_ "github.com/mattn/go-sqlite3"
	"github.com/powerman/structlog"
	"net/rpc"
	"sync"
)

type App struct{}

func Run() error {

	// Преверяем, не было ли приложение запущено ранее.
	// Если было, выдвигаем окно UI приложения на передний план и завершаем процесс.
	if notify.ServerWindowAlreadyExists {
		hWnd := winapp.FindWindow(notify.PeerWindowClassName)
		win.ShowWindow(hWnd, win.SW_RESTORE)
		win.SetForegroundWindow(hWnd)
		log.Fatal("elco.exe already executing")
	}

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

	peer.Init()

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

	notify.Window.Close()

	winapp.EnumWindowsWithClassName(func(hWnd win.HWND, winClassName string) {
		if winClassName == notify.PeerWindowClassName {
			r := win.PostMessage(hWnd, win.WM_CLOSE, 0, 0)
			log.Debug("close peer window", "syscall", r)
		}
	})

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
