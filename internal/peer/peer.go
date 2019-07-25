package peer

import (
	"github.com/fpawel/gohelp/copydata"
	"github.com/fpawel/gohelp/winapp"
	"github.com/getlantern/systray"
	"github.com/lxn/win"
	"github.com/powerman/structlog"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	ServerWindowClassName = "ElcoServerWindow"
	WindowClassName       = "TElcoMainForm"
)

func Notifyf(msg uintptr, format string, a ...interface{}) {
	if notifyWindow == nil {
		panic("not initialized")
	}
	notifyWindow.Notifyf(msg, format, a...)
}

func NotifyStr(msg uintptr, s string) {
	if notifyWindow == nil {
		panic("not initialized")
	}
	notifyWindow.NotifyStr(msg, s)
}

func NotifyJson(msg uintptr, param interface{}) {
	if notifyWindow == nil {
		panic("not initialized")
	}
	notifyWindow.NotifyJson(msg, param)
}

func Close() {
	if notifyWindow == nil {
		panic("already closed")
	}
	notifyWindow.Close()
	winapp.EnumWindowsWithClassName(func(hWnd win.HWND, winClassName string) {
		if winClassName == WindowClassName {
			r := win.PostMessage(hWnd, win.WM_CLOSE, 0, 0)
			log.Debug("close peer window", "syscall", r)
		}
	})
	notifyWindow = nil

}

func ResetPeer() {
	if notifyWindow == nil {
		panic("not initialized")
	}
	notifyWindow.ResetPeer()
}

func InitPeer() {
	if notifyWindow == nil {
		panic("not initialized")
	}
	notifyWindow.InitPeer()
}

func InitNotifyWindow(serverWindowClassNameSuffix string) {
	notifyWindow = copydata.NewNotifyWindow(ServerWindowClassName+serverWindowClassNameSuffix,
		WindowClassName)
}

func Init(serverWindowClassNameSuffix string) {
	if notifyWindow != nil {
		panic("already init")
	}

	// Преверяем, не было ли приложение запущено ранее.
	// Если было, выдвигаем окно UI приложения на передний план и завершаем процесс.
	if winapp.IsWindow(winapp.FindWindow(ServerWindowClassName)) {
		hWnd := winapp.FindWindow(WindowClassName)
		win.ShowWindow(hWnd, win.SW_RESTORE)
		win.SetForegroundWindow(hWnd)
		log.Fatal("elco.exe already executing")
	}
	InitNotifyWindow(serverWindowClassNameSuffix)

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
					systray.Quit()
					notifyWindow.Close()
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
	// окно для отправки сообщений WM_COPYDATA дельфи-приложению
	notifyWindow *copydata.NotifyWindow
)
