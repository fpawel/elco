package daemon

import (
	"context"
	"github.com/Microsoft/go-winio"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/assets"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/elco/internal/ktx500"
	"github.com/fpawel/elco/pkg/copydata"
	"github.com/fpawel/elco/pkg/winapp"
	"github.com/getlantern/systray"
	"github.com/jmoiron/sqlx"
	"github.com/lxn/win"
	_ "github.com/mattn/go-sqlite3"
	"github.com/powerman/rpc-codec/jsonrpc2"
	"github.com/powerman/structlog"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/sqlite3"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type D struct {
	dbProducts   *reform.DB
	dbxProducts  *sqlx.DB
	notifyWindow *copydata.NotifyWindow // окно для отправки сообщений WM_COPYDATA дельфи-приложению
	cfg          *cfg.Config
	ctx          context.Context
	hardware     hardware

	log *structlog.Logger

	portMeasurer  *comport.Reader
	portGas       *comport.Reader
	muCurrentWork sync.Mutex
}

type hardware struct {
	sync.WaitGroup
	Continue,
	cancel,
	skipDelay context.CancelFunc
	ctx context.Context
}

func Run(skipRunUIApp, createNewDB bool) error {

	dbProductsConn, err := data.Open(createNewDB)
	if err != nil {
		return merry.WithMessage(err, "не удалось открыть файл данных")
	}

	x := &D{
		log:         structlog.New(),
		dbProducts:  reform.NewDB(dbProductsConn, sqlite3.Dialect, nil),
		dbxProducts: sqlx.NewDb(dbProductsConn, "sqlite3"),
		hardware: hardware{
			Continue:  func() {},
			cancel:    func() {},
			skipDelay: func() {},
			ctx:       context.Background(),
		},
	}

	x.notifyWindow = copydata.NewNotifyWindow(
		elco.ServerWindowClassName,
		elco.PeerWindowClassName,
		x.log, notify.FormatMsg)

	x.cfg = cfg.OpenConfig(x.dbProducts)

	x.portMeasurer = comport.NewReader(comport.Config{
		Baud:        115200,
		ReadTimeout: time.Millisecond,
	})

	x.portGas = comport.NewReader(comport.Config{
		Baud:        9600,
		ReadTimeout: time.Millisecond,
	})

	go runSysTray(x.notifyWindow.CloseWindow)

	for _, svcObj := range []interface{}{
		api.NewPartiesCatalogue(x.dbProducts, x.dbxProducts),
		api.NewLastParty(x.dbProducts, x.dbxProducts),
		api.NewProductTypes(x.dbProducts),
		api.NewProductFirmware(x.dbProducts, x),
		api.NewSetsSvc(x.cfg),
		&api.RunnerSvc{Runner: x},
	} {
		if err := rpc.Register(svcObj); err != nil {
			return merry.Wrap(err)
		}
	}

	var cancel func()
	x.ctx, cancel = context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	wg.Add(1)

	ln := mustPipeListener()

	// цикл RPC сервера
	go func() {
		defer wg.Done()
		defer x.notifyWindow.CloseWindow()
		x.serveRPC(ln, x.ctx)
	}()

	if !skipRunUIApp {
		if err := runUIApp(); err != nil {
			return merry.WithMessage(err, "не удалось выполнить elcoui.exe")
		}
	} else {
		x.log.Debug("elcoui.exe не будет запущен, поскольку установлен соответствующий флаг")
	}
	notify.StartServerApplication(x.notifyWindow, "")

	ktx500.GetConfig = func() cfg.FinsNetwork {
		return x.cfg.Predefined().FinsNetwork
	}

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

	x.log.ErrIfFail(ln.Close, "defer", "close pipe listener")

	wg.Wait()

	winapp.EnumWindowsWithClassName(func(hWnd win.HWND, winClassName string) {
		if winClassName == elco.PeerWindowClassName {
			r := win.PostMessage(hWnd, win.WM_CLOSE, 0, 0)
			x.log.Debug("close peer window", "syscall", r)
		}
	})
	x.log.ErrIfFail(dbProductsConn.Close, "defer", "close products db")
	x.log.ErrIfFail(x.cfg.Save, "defer", "save config")
	return nil
}

func (x *D) serveRPC(ln net.Listener, ctx context.Context) {

	for {
		switch conn, err := ln.Accept(); err {
		case nil:
			go jsonrpc2.ServeConnContext(ctx, conn)
		case winio.ErrPipeListenerClosed:
			x.log.Debug("rpc pipe was closed")
			return
		default:
			x.log.PrintErr(merry.Append(err, "rpc pipe error"))
			return
		}
	}
}

func mustPipeListener() net.Listener {
	ln, err := winio.ListenPipe(elco.PipeName, nil)
	if err != nil {
		panic(err)
	}
	return ln
}

func runUIApp() error {
	fileName := filepath.Join(filepath.Dir(os.Args[0]), "elcoui.exe")
	err := exec.Command(fileName).Start()
	if err != nil {
		return merry.Append(err, fileName)
	}
	return nil
}

func runSysTray(onClose func()) {
	systray.Run(func() {

		appIconBytes, err := assets.Asset("assets/appicon.ico")
		if err != nil {
			panic(err)
		}

		systray.SetIcon(appIconBytes)
		systray.SetTitle("Производство ЭХЯ CO")
		systray.SetTooltip("Производство ЭХЯ CO")
		mRunGUIApp := systray.AddMenuItem("Показать", "Показать окно приложения")
		mQuitOrig := systray.AddMenuItem("Закрыть", "Закрыть приложение")

		go func() {
			for {
				select {
				case <-mRunGUIApp.ClickedCh:
					if err := runUIApp(); err != nil {
						panic(merry.Append(err, "не удалось запустить elcoui.exe"))
					}
				case <-mQuitOrig.ClickedCh:
					systray.Quit()
					onClose()
				}
			}
		}()
	}, func() {
	})
}
