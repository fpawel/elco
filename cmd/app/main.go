package main

import (
	"flag"
	"github.com/fpawel/elco/internal/daemon"
	"github.com/fpawel/goutils/winapp"
	"github.com/lxn/win"
	"github.com/sirupsen/logrus"
)

func main() {

	// Преверяем, не было ли приложение запущено ранее
	if winapp.IsWindow(winapp.FindWindow(daemon.ServerWindowClassName)) {
		// Если было, выдвигаем окно приложения на передний план и завершаем процесс
		hWnd := winapp.FindWindow(daemon.PeerWindowClassName)
		win.ShowWindow(hWnd, win.SW_RESTORE)
		win.SetForegroundWindow(hWnd)
		return
	}

	mustRunPeer := true
	flag.BoolVar(&mustRunPeer, "must-run-peer", true, "ensure peer application")

	logLevelStr := "warn"
	flag.StringVar(&logLevelStr, "log-level", "warn", "use log level")

	flag.Parse()

	// Log as JSON instead of the default ASCII formatter.
	//logrus.SetFormatter(&logrus.JSONFormatter{})

	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		logrus.Fatal(err)
	}
	// Only log the warning severity or above.
	logrus.SetLevel(logLevel)
	logrus.SetReportCaller(true)

	d := daemon.New()
	d.Run(mustRunPeer)

	if err := d.Close(); err != nil {
		panic(err)
	}

	if mustRunPeer {
		if err := closeAllPeerWindows(); err != nil {
			panic(err)
		}
	}
}
