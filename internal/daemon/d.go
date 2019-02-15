package daemon

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/Microsoft/go-winio"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/elco/pkg/copydata"
	"github.com/fpawel/elco/pkg/serial-comm/comport"
	"github.com/fpawel/elco/pkg/winapp"
	"github.com/lxn/walk"
	"github.com/lxn/win"
	_ "github.com/mattn/go-sqlite3"
	"github.com/powerman/rpc-codec/jsonrpc2"
	"github.com/sirupsen/logrus"
	"gopkg.in/reform.v1"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type D struct {
	db  *reform.DB
	w   *copydata.NotifyWindow // окно для отправки сообщений WM_COPYDATA дельфи-приложению
	cfg *data.Config
	ctx context.Context

	port struct {
		measurer, gas *comport.Port
	}

	hardware hardware
}

type hardware struct {
	sync.WaitGroup
	Continue,
	cancel,
	skipDelay context.CancelFunc
	ctx       context.Context
	logFields logFields
}

type logFields logrus.Fields

func (x logFields) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (x logFields) Fire(entry *logrus.Entry) error {
	for k, v := range x {
		entry.Data[k] = v
	}
	return nil
}

func Run(skipRunUIApp bool) {

	dbConn, err := sql.Open("sqlite3", elco.DataFileName())
	if err != nil {
		logrus.Fatalln("Не удалось открыть основной файл данных:", err, elco.DataFileName())
	}
	dbConn.SetMaxIdleConns(1)
	dbConn.SetMaxOpenConns(1)
	dbConn.SetConnMaxLifetime(0)

	// reform.NewPrintfLogger(logrus.Debugf)
	db, err := data.Open(dbConn, nil)
	if err != nil {
		logrus.Fatalln("Не удалось открыть основной файл данных:", err, elco.DataFileName())
	}
	x := &D{
		db:  db,
		cfg: data.OpenConfig(db),
		w:   copydata.NewNotifyWindow(elco.ServerWindowClassName, elco.PeerWindowClassName),
		hardware: hardware{
			Continue:  func() {},
			cancel:    func() {},
			skipDelay: func() {},
			ctx:       context.Background(),
			logFields: make(logFields),
		},
	}
	x.port.measurer = comport.NewPortWithHook(x.onComport)
	x.port.gas = comport.NewPortWithHook(x.onComport)
	logrus.AddHook(&x.hardware.logFields)

	notifyIcon, err := walk.NewNotifyIcon()
	if err != nil {
		logrus.Panicln("unable to create the notify icon:", err)
	}
	setupNotifyIcon(notifyIcon, x.w.CloseWindow)

	for _, svcObj := range []interface{}{
		api.NewPartiesCatalogue(x.db),
		api.NewLastParty(x.db),
		api.NewProductTypes(x.db),
		api.NewProductFirmware(x.db, x),
		api.NewSetsSvc(x.cfg),
		&api.RunnerSvc{Runner: x},
	} {
		if err := rpc.Register(svcObj); err != nil {
			panic(err)
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
		defer x.w.CloseWindow()
		x.serveRPC(ln, x.ctx)
	}()

	if !skipRunUIApp {
		if err := runUIApp(); err != nil {
			logrus.Panicln(err)
		}
	} else {
		logrus.Warn("skip running ui flag set")
	}
	notify.StartServerApplication(x.w, "")

	// цикл оконных сообщений
	for {
		var msg win.MSG
		if win.GetMessage(&msg, 0, 0, 0) == 0 {
			break
		}
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}

	x.hardware.cancel()

	cancel()

	logrus.Debugln("close pipe listener on exit:", ln.Close())

	wg.Wait()

	logrus.Debugln("clean up notify icon on exit:", notifyIcon.Dispose())

	winapp.EnumWindowsWithClassName(func(hWnd win.HWND, winClassName string) {
		if winClassName == elco.PeerWindowClassName {
			logrus.Debugln("close peer window:", hWnd, winClassName)
			win.PostMessage(hWnd, win.WM_CLOSE, 0, 0)
		}
	})
	logrus.Infoln("close sqlite data base on exit:", dbConn.Close())
	logrus.Infoln("save config on exit:", x.cfg.Save())
}

func (x *D) onComport(w comport.LastWork) {
	if x.cfg.User().LogComports {
		notify.ComportEntry(x.w, api.ComportEntry{
			Port:  w.Port,
			Error: w.Error != nil,
			Msg:   w.String(),
		})
	}
}

func (x *D) serveRPC(ln net.Listener, ctx context.Context) {

	for {
		switch conn, err := ln.Accept(); err {
		case nil:
			go jsonrpc2.ServeConnContext(ctx, conn)
		case winio.ErrPipeListenerClosed:
			return
		default:
			fmt.Println("rpc pipe error:", err)
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
	const (
		peerAppExe = "elcoui.exe"
	)
	dir := filepath.Dir(os.Args[0])

	if _, err := os.Stat(filepath.Join(dir, peerAppExe)); os.IsNotExist(err) {
		dir = elco.AppName.Dir()
	}

	cmd := exec.Command(filepath.Join(dir, peerAppExe))
	cmd.Stdout = os.Stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	return cmd.Start()
}

func setupNotifyIcon(notifyIcon *walk.NotifyIcon, exitFunc func()) {
	iconBytes, err := FSByte(false, "/img/appicon.ico")
	if err != nil {
		logrus.Panicf("unable to restore elco icon bytes: %v", err)
	}

	//We load our icon from a temp file.
	appIcon, err := winapp.IconFromBytes(iconBytes)
	if err != nil {
		log.Panicln(err)
	}

	// Set the icon and a tool tip text.
	if err := notifyIcon.SetIcon(appIcon); err != nil {
		log.Panicln(err)
	}

	if err := notifyIcon.SetToolTip("Производство ЭХЯ CO"); err != nil {
		logrus.Panic(err)
	}

	// When the left mouse button is pressed, bring up our balloon.
	notifyIcon.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button != walk.LeftButton {
			return
		}
		logrus.Debugln("sys tray: clicked")
		if err := runUIApp(); err != nil {
			logrus.Panicln("unable to run ui elco: ", err)
		}
	})

	// We put an exit action into the context menu.
	exitAction := walk.NewAction()
	if err := exitAction.SetText("Выход"); err != nil {
		logrus.Panic(err)
	}
	exitAction.Triggered().Attach(func() {
		logrus.Debugln("sys tray: \"Выход\" clicked")
		exitFunc()
	})

	if err := notifyIcon.ContextMenu().Actions().Add(exitAction); err != nil {
		logrus.Panic(err)
	}

	// The notify icon is hidden initially, so we have to make it visible.
	if err := notifyIcon.SetVisible(true); err != nil {
		logrus.Panic(err)
	}
}
