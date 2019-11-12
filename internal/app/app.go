package app

import (
	"context"
	"fmt"
	"github.com/fpawel/elco/internal"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/pkg/ktx500"
	"github.com/fpawel/elco/internal/pkg/must"
	"github.com/fpawel/elco/internal/pkg/winapp"
	"github.com/lxn/win"
	_ "github.com/mattn/go-sqlite3"
	"github.com/powerman/structlog"
	"net/rpc"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
)

type App struct{}

func Run() error {
	// Если было приложение запущено ранее, завершить процесс.
	if winapp.IsWindow(winapp.FindWindow(internal.WindowClassName)) {
		log.Fatal("elco.exe already executing")
	}

	data.Open()

	var interrupt func()
	ctxApp, interrupt = context.WithCancel(context.Background())

	for _, svcObj := range []interface{}{
		new(api.PartiesCatalogueSvc),
		new(api.LastPartySvc),
		new(api.ProductTypesSvc),
		api.NewProductFirmware(runner{}),
		new(api.PdfSvc),
		&api.RunnerSvc{Runner: runner{}},
		new(api.ConfigSvc),
		new(api.ProductsCatalogueSvc),
	} {
		must.AbortIf(rpc.Register(svcObj))
	}

	closeHttpServer := startHttpServer()

	// run GUI
	go func() {
		defer interrupt()

		runtime.LockOSThread()

		// инициализация окна окно связи с GUI для отправки сообщений WM_COPYDATA
		if !winapp.IsWindow(winapp.NewWindowWithClassName(internal.WindowClassName, win.DefWindowProc)) {
			panic("window was not created")
		}

		for {
			var msg win.MSG
			if win.GetMessage(&msg, 0, 0, 0) == 0 {
				log.Info("выход из цикла оконных сообщений")
				return
			}
			log.Debug(fmt.Sprintf("%+v", msg))
			win.TranslateMessage(&msg)
			win.DispatchMessage(&msg)
		}
	}()

	go ktx500.TraceTemperature(func(info api.Ktx500Info) {
		notify.Ktx500Info(nil, info)
	}, func(s string) {
		notify.Ktx500Error(nil, s)
	})

	if len(os.Getenv("ELCO_DEV_MODE")) != 0 {
		log.Debug("waiting system signal because of ELCO_DEV_MODE=" + os.Getenv("ELCO_DEV_MODE"))
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		sig := <-done
		log.Debug("system signal: " + sig.String())
	} else {
		cmd := exec.Command(filepath.Join(filepath.Dir(os.Args[0]), "elcoui.exe"))
		log.ErrIfFail(cmd.Start)
		log.ErrIfFail(cmd.Wait)
		log.Debug("gui was closed.")
	}

	interrupt()
	closeHttpServer()
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
