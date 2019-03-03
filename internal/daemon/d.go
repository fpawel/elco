package daemon

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Microsoft/go-winio"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/elco/internal/journal"
	"github.com/fpawel/elco/pkg/copydata"
	"github.com/fpawel/elco/pkg/serial-comm/comport"
	"github.com/fpawel/elco/pkg/winapp"
	"github.com/go-logfmt/logfmt"
	"github.com/jmoiron/sqlx"
	"github.com/lxn/walk"
	"github.com/lxn/win"
	_ "github.com/mattn/go-sqlite3"
	"github.com/powerman/rpc-codec/jsonrpc2"
	"github.com/sirupsen/logrus"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/sqlite3"
	"net"
	"net/rpc"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type D struct {
	dbProducts    *reform.DB
	dbxProducts   *sqlx.DB
	dbJournal     *reform.DB
	w             *copydata.NotifyWindow // окно для отправки сообщений WM_COPYDATA дельфи-приложению
	cfg           *cfg.Config
	ctx           context.Context
	hardware      hardware
	logFields     logrus.Fields
	portMeasurer  *comport.Port
	portGas       *comport.Port
	muCurrentWork sync.Mutex
	currentWork   *journal.Work
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

	dbJournalConn, err := journal.Open(createNewDB)
	if err != nil {
		return merry.WithMessage(err, "не удалось открыть журнал")
	}

	x := &D{
		dbProducts:  reform.NewDB(dbProductsConn, sqlite3.Dialect, nil),
		dbxProducts: sqlx.NewDb(dbProductsConn, "sqlite3"),
		dbJournal:   reform.NewDB(dbJournalConn, sqlite3.Dialect, nil),
		w:           copydata.NewNotifyWindow(elco.ServerWindowClassName, elco.PeerWindowClassName),
		hardware: hardware{
			Continue:  func() {},
			cancel:    func() {},
			skipDelay: func() {},
			ctx:       context.Background(),
		},
		logFields: make(logrus.Fields),
	}
	if err := data.Init(x.dbProducts); err != nil {
		return merry.Wrap(err)
	}
	x.cfg = cfg.OpenConfig(x.dbProducts)
	if err := journal.Init(x.dbJournal); err != nil {
		return merry.Wrap(err)
	}
	elco.Logger.AddHook(x)
	logrus.AddHook(x)
	x.portMeasurer = comport.NewPort("стенд", x.onComport)
	x.portGas = comport.NewPort("пневмоблок", x.onComport)

	notifyIcon, err := walk.NewNotifyIcon()
	if err != nil {
		return merry.WithMessage(err, "unable to create the notify icon")
	}

	if err := setupNotifyIcon(notifyIcon, x.w.CloseWindow); err != nil {
		return merry.WithMessage(err, "unable to create the notify icon")
	}

	for _, svcObj := range []interface{}{
		api.NewPartiesCatalogue(x.dbProducts, x.dbxProducts),
		api.NewLastParty(x.dbProducts, x.dbxProducts),
		api.NewProductTypes(x.dbProducts),
		api.NewProductFirmware(x.dbProducts, x),
		api.NewSetsSvc(x.cfg),
		api.NewJournal(x.dbJournal, sqlx.NewDb(dbJournalConn, "sqlite3")),
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
		defer x.w.CloseWindow()
		x.serveRPC(ln, x.ctx)
	}()

	if !skipRunUIApp {
		if err := runUIApp(); err != nil {
			return merry.WithMessage(err, "unable to execute gui application")
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
	logrus.Infoln("close products data base on exit:", dbProductsConn.Close())
	logrus.Infoln("close journal data base on exit:", dbJournalConn.Close())
	logrus.Infoln("save config on exit:", x.cfg.Save())
	return nil
}

func (x *D) onComport(w comport.Entry) {
	if x.cfg.Predefined().VerboseLogging {
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
			logrus.Errorln("rpc pipe error:", err)
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
	fileName, err := winapp.CurrentDirOrProfileFileName(".elco", "elcoui.exe")
	if err != nil {
		return merry.Wrap(err)
	}
	return exec.Command(fileName).Start()
}

func setupNotifyIcon(notifyIcon *walk.NotifyIcon, exitFunc func()) error {

	appIconFileName, err := winapp.ProfileFileName(".elco", "assets", "appicon.ico")
	if err != nil {
		return merry.Wrap(err)
	}

	//We load our icon from a temp file.
	appIcon, err := walk.NewIconFromFile(appIconFileName)
	if err != nil {
		return merry.Wrap(err)
	}

	// Set the icon and a tool tip text.
	if err := notifyIcon.SetIcon(appIcon); err != nil {
		return merry.Wrap(err)
	}

	if err := notifyIcon.SetToolTip("Производство ЭХЯ CO"); err != nil {
		return merry.Wrap(err)
	}

	// When the left mouse button is pressed, bring up our balloon.
	notifyIcon.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button != walk.LeftButton {
			return
		}
		logrus.Debugln("sys tray: clicked")
		if err := runUIApp(); err != nil {
			logrus.Panicln("unable to run gui aplication elcoui.exe:", err)
		}
	})

	// We put an exit action into the context menu.
	exitAction := walk.NewAction()
	if err := exitAction.SetText("Выход"); err != nil {
		return merry.Wrap(err)
	}
	exitAction.Triggered().Attach(func() {
		logrus.Debugln("sys tray: \"Выход\" clicked")
		exitFunc()
	})

	if err := notifyIcon.ContextMenu().Actions().Add(exitAction); err != nil {
		return merry.Wrap(err)
	}

	// The notify icon is hidden initially, so we have to make it visible.
	if err := notifyIcon.SetVisible(true); err != nil {
		return merry.Wrap(err)
	}
	return nil
}

func (x *D) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (x *D) Fire(entry *logrus.Entry) error {

	for k, v := range x.logFields {
		if len(entry.Data) == 0 {
			entry.Data = logrus.Fields{}
		}
		entry.Data[k] = v
	}
	x.muCurrentWork.Lock()
	currentWork := x.currentWork
	x.muCurrentWork.Unlock()
	if currentWork == nil {
		return nil
	}
	c := entry.Caller

	journalEntry := journal.Entry{
		Message:   entry.Message,
		Level:     entry.Level.String(),
		WorkID:    currentWork.WorkID,
		CreatedAt: time.Now(),
		Line:      int64(c.Line),
		File:      filepath.Base(c.File),
	}

	sb := bytes.NewBuffer(nil)
	d := logfmt.NewEncoder(sb)
	for k, v := range entry.Data {
		switch k {
		case "stack":
			journalEntry.Stack = fmt.Sprintf("%v", v)
		case "msg", "message", "time", "level", "work":
		default:
			_ = d.EncodeKeyval(k, v)
		}
	}
	if sb.Len() > 0 {
		if len(journalEntry.Message) > 0 {
			journalEntry.Message += " "
		}
		journalEntry.Message += sb.String()
	}
	err := x.dbJournal.Save(&journalEntry)
	entry.Data["entry_id"] = journalEntry.EntryID
	notify.NewJournalEntry(x.w, journalEntry.EntryInfo(x.currentWork.Name))
	return err
}
