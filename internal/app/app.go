package app

import (
	"context"
	"github.com/fpawel/elco/internal"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/peer"
	"github.com/fpawel/elco/internal/pkg/ktx500"
	"github.com/fpawel/elco/internal/pkg/must"
	"github.com/fpawel/elco/internal/pkg/winapp"
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
	if winapp.IsWindow(winapp.FindWindow(internal.ServerWindowClassName)) {
		hWnd := winapp.FindWindow(internal.PeerWindowClassName)
		win.ShowWindow(hWnd, win.SW_RESTORE)
		win.SetForegroundWindow(hWnd)
		log.Fatal("elco.exe already executing")
	}
	winapp.NewWindowWithClassName(internal.ServerWindowClassName, win.DefWindowProc)

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

	go ktx500.TraceTemperature(func(info api.Ktx500Info) {
		notify.Ktx500Info(nil, info)
	}, func(s string) {
		notify.Ktx500Error(nil, s)
	})

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

	win.PostMessage(winapp.FindWindow(internal.ServerWindowClassName), win.WM_CLOSE, 0, 0)

	winapp.EnumWindowsWithClassName(func(hWnd win.HWND, winClassName string) {
		if winClassName == internal.PeerWindowClassName {
			win.PostMessage(hWnd, win.WM_CLOSE, 0, 0)
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
