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
	"github.com/fpawel/elco/internal/data/journal"
	"github.com/fpawel/elco/internal/elco"
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
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type D struct {
	db, dbJournal   *reform.DB
	dbx, dbxJournal *sqlx.DB
	w               *copydata.NotifyWindow // окно для отправки сообщений WM_COPYDATA дельфи-приложению
	cfg             *data.Config
	ctx             context.Context
	hardware        hardware
	logFields       logrus.Fields
	portMeasurer    *comport.Port
	portGas         *comport.Port
	muCurrentWork   sync.Mutex
	currentWork     *journal.Work
}

type hardware struct {
	sync.WaitGroup
	Continue,
	cancel,
	skipDelay context.CancelFunc
	ctx context.Context
}

func Run(skipRunUIApp bool) {

	dbConn, err := sql.Open("sqlite3", elco.DataFileName())
	if err != nil {
		logrus.Fatalln("Не удалось открыть основной файл данных:", err, elco.DataFileName())
	}
	dbConn.SetMaxIdleConns(1)
	dbConn.SetMaxOpenConns(1)
	dbConn.SetConnMaxLifetime(0)

	dbJournalConn, err := sql.Open("sqlite3", elco.JournalFileName())
	if err != nil {
		logrus.Fatalln("Не удалось открыть журнал:", err, elco.JournalFileName())
	}
	dbJournalConn.SetMaxIdleConns(1)
	dbJournalConn.SetMaxOpenConns(1)
	dbJournalConn.SetConnMaxLifetime(0)

	dbJournal, err := journal.Open(dbJournalConn, nil)
	if err != nil {
		logrus.Fatalln("Не удалось открыть журнал:", err, elco.JournalFileName())
	}

	// reform.NewPrintfLogger(logrus.Debugf)
	db, err := data.Open(dbConn, nil)
	if err != nil {
		logrus.Fatalln("Не удалось открыть основной файл данных:", err, elco.DataFileName())
	}
	x := &D{
		db:         db,
		dbx:        sqlx.NewDb(dbConn, "sqlite3"),
		dbJournal:  dbJournal,
		dbxJournal: sqlx.NewDb(dbJournalConn, "sqlite3"),
		cfg:        data.OpenConfig(db),
		w:          copydata.NewNotifyWindow(elco.ServerWindowClassName, elco.PeerWindowClassName),
		hardware: hardware{
			Continue:  func() {},
			cancel:    func() {},
			skipDelay: func() {},
			ctx:       context.Background(),
		},
		logFields: make(logrus.Fields),
	}
	elco.Logger.AddHook(x)
	logrus.AddHook(x)
	x.portMeasurer = comport.NewPort(logrus.Fields{"device": "стенд"}, x.onComport)
	x.portGas = comport.NewPort(logrus.Fields{"device": "пневмоблок"}, x.onComport)

	notifyIcon, err := walk.NewNotifyIcon()
	if err != nil {
		logrus.Panicln("unable to create the notify icon:", err)
	}
	setupNotifyIcon(notifyIcon, x.w.CloseWindow)

	for _, svcObj := range []interface{}{
		api.NewPartiesCatalogue(x.db, x.dbx),
		api.NewLastParty(x.db),
		api.NewProductTypes(x.db),
		api.NewProductFirmware(x.db, x),
		api.NewSetsSvc(x.cfg),
		api.NewJournal(x.dbJournal, x.dbxJournal),
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

	sb := bytes.NewBuffer(nil)
	d := logfmt.NewEncoder(sb)
	excludeFields := map[string]struct{}{
		"msg":     {},
		"message": {},
		"time":    {},
		"level":   {},
		"work":    {},
		"stack":   {},
	}

	for k, v := range entry.Data {
		if _, f := excludeFields[k]; !f {
			_ = d.EncodeKeyval(k, v)
		}
	}

	journalEntry := journal.Entry{
		Message:   entry.Message,
		Level:     entry.Level.String(),
		WorkID:    currentWork.WorkID,
		CreatedAt: time.Now(),
	}
	if sb.Len() > 0 {
		if len(journalEntry.Message) > 0 {
			journalEntry.Message += " "
		}
		journalEntry.Message += sb.String()
	}

	c := entry.Caller
	journalEntry.Message += fmt.Sprintf(" %s:%d", filepath.Base(c.File), c.Line)

	err := x.dbJournal.Save(&journalEntry)
	entry.Data["entry_id"] = journalEntry.EntryID

	notify.NewJournalEntry(x.w, journal.EntryInfo{
		CreatedAt: journalEntry.CreatedAt,
		EntryID:   journalEntry.EntryID,
		WorkID:    currentWork.WorkID,
		WorkName:  currentWork.Name,
		Message:   journalEntry.Message,
		Level:     journalEntry.Level,
	})
	return err
}
