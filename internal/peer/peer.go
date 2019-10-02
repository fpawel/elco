package peer

import (
	"github.com/fpawel/elco/internal"
	"github.com/fpawel/elco/internal/pkg/winapp"
	"github.com/getlantern/systray"
	"github.com/lxn/win"
	"github.com/powerman/structlog"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

func Init() {
	if os.Getenv("ELCO_SKIP_RUN_PEER") == "true" {
		log.Warn("ELCO_SKIP_RUN_PEER")
	} else {
		show()
	}

	go systray.Run(func() {

		appIconBytes, err := ioutil.ReadFile(filepath.Join(filepath.Dir(os.Args[0]), "assets", "appicon.ico"))
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
					show()
				case <-mQuitOrig.ClickedCh:
					win.PostMessage(winapp.FindWindow(internal.ServerWindowClassName), win.WM_CLOSE, 0, 0)
					systray.Quit()
				}
			}
		}()
	}, func() {
	})
}

func show() {
	if err := exec.Command(filepath.Join(filepath.Dir(os.Args[0]), "elcoui.exe")).Start(); err != nil {
		panic(err)
	}
}

var (
	log = structlog.New()
)
