package daemon

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Microsoft/go-winio"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/assets"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/elco/internal/journal"
	"github.com/fpawel/elco/pkg/copydata"
	"github.com/fpawel/elco/pkg/winapp"
	"github.com/getlantern/systray"
	"github.com/go-logfmt/logfmt"
	"github.com/jmoiron/sqlx"
	"github.com/lxn/win"
	_ "github.com/mattn/go-sqlite3"
	"github.com/powerman/rpc-codec/jsonrpc2"
	"github.com/powerman/structlog"
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
	dbProducts  *reform.DB
	dbxProducts *sqlx.DB
	dbJournal   *reform.DB
	w           *copydata.NotifyWindow // окно для отправки сообщений WM_COPYDATA дельфи-приложению
	cfg         *cfg.Config
	ctx         context.Context
	hardware    hardware

	log *structlog.Logger

	portMeasurer  *comport.Reader
	portGas       *comport.Reader
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
		log:         structlog.New(),
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
	}
	x.cfg = cfg.OpenConfig(x.dbProducts)

	elco.Logger.AddHook(x)
	logrus.AddHook(x)

	x.portMeasurer = comport.NewReader(comport.Config{
		Baud:        115200,
		ReadTimeout: time.Millisecond,
	})

	x.portGas = comport.NewReader(comport.Config{
		Baud:        9600,
		ReadTimeout: time.Millisecond,
	})

	go runSysTray(x.w.CloseWindow)

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

func (x *D) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (x *D) Fire(entry *logrus.Entry) error {

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

func runSysTray(onClose func()) {
	systray.Run(func() {

		appIconBytes, err := assets.Asset("assets/appicon.ico")
		if err != nil {
			logrus.Fatal(err)
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
						logrus.Panicln("unable to run gui aplication elcoui.exe:", err)
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
