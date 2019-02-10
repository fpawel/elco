package daemon

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Microsoft/go-winio"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/crud"
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/elco/internal/elco/config"
	"github.com/fpawel/goutils/copydata"
	"github.com/fpawel/goutils/serial-comm/comport"
	"github.com/fpawel/goutils/winapp"
	"github.com/lxn/walk"
	"github.com/lxn/win"
	"github.com/powerman/rpc-codec/jsonrpc2"
	"github.com/sirupsen/logrus"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type D struct {
	c    crud.DBContext         // база данных sqlite
	w    *copydata.NotifyWindow // окно для отправки сообщений WM_COPYDATA дельфи-приложению
	sets *config.Sets
	ctx  context.Context

	port struct {
		measurer, gas *comport.Port
	}

	hardware struct {
		sync.WaitGroup
		Continue,
		cancel,
		skipDelay context.CancelFunc
		ctx       context.Context
		logFields logFields
	}
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

func New() *D {

	// reform.NewPrintfLogger(logrus.Debugf)
	c := crud.NewDBContext(nil)
	sets := config.OpenSets(c.LastParty())
	x := &D{
		c:    c,
		sets: sets,
		w:    copydata.NewNotifyWindow(elco.ServerWindowClassName, elco.PeerWindowClassName),
	}
	x.port.measurer = new(comport.Port)
	x.port.gas = new(comport.Port)
	x.hardware.cancel = func() {}
	x.hardware.Continue = func() {}
	x.hardware.skipDelay = func() {}
	x.hardware.ctx = context.Background()
	x.hardware.logFields = make(logFields)
	logrus.AddHook(&x.hardware.logFields)

	return x
}

func (x *D) Run(skipRunUIApp bool) {

	notifyIcon, err := walk.NewNotifyIcon()
	if err != nil {
		logrus.Panicln("unable to create the notify icon:", err)
	}
	setupNotifyIcon(notifyIcon, x.w.CloseWindow)

	for _, svcObj := range []interface{}{
		api.NewPartiesCatalogue(x.c.PartiesCatalogue()),
		api.NewLastParty(x.c.LastParty()),
		api.NewProductTypes(x.c.ProductTypes()),
		api.NewProductFirmware(x.c.ProductFirmware()),
		api.NewSetsSvc(x.sets),
		&api.RunnerSvc{x},
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

	logrus.Debugln("close sqlite data base on exit:", x.c.Close())
	logrus.Debugln("save config on exit:", x.sets.Save())
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
