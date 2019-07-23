package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/assets"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/elco/internal/ktx500"
	"github.com/fpawel/gohelp/winapp"
	"github.com/getlantern/systray"
	"github.com/lxn/win"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"os/exec"
	"path/filepath"
)

type App struct{}

func Run(skipRunUIApp, createNewDB bool) error {

	cfg.OpenConfig()
	data.Open(createNewDB)
	notify.InitWindow("")
	closeHttpServer := startHttpServer()
	go runSysTray(notify.W.Close)

	if !skipRunUIApp {
		if err := runUIApp(); err != nil {
			return merry.WithMessage(err, "не удалось выполнить elcoui.exe")
		}
	} else {
		log.Debug("elcoui.exe не будет запущен, поскольку установлен соответствующий флаг")
	}

	var cancel func()
	ctxApp, cancel = context.WithCancel(context.Background())

	notify.StartServerApplication(log, "")

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

	winapp.EnumWindowsWithClassName(func(hWnd win.HWND, winClassName string) {
		if winClassName == elco.PeerWindowClassName {
			r := win.PostMessage(hWnd, win.WM_CLOSE, 0, 0)
			log.Debug("close peer window", "syscall", r)
		}
	})
	log.ErrIfFail(data.Close, "defer", "close products db")
	log.ErrIfFail(cfg.Cfg.Save, "defer", "save config")
	return nil
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