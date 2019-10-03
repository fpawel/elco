package app

import (
	"github.com/fpawel/elco/internal"
	"github.com/getlantern/systray"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

func runGUI() {
	show := func() {
		if err := exec.Command(filepath.Join(filepath.Dir(os.Args[0]), "elcoui.exe")).Start(); err != nil {
			panic(err)
		}
	}
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
					internal.CloseHWnd()
					systray.Quit()
				}
			}
		}()
	}, internal.CloseHWnd)
}
