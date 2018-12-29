package main

import (
	"flag"
	"github.com/fpawel/elco/internal/app"
	"github.com/fpawel/elco/internal/daemon"
	"github.com/fpawel/goutils/winapp"
	"github.com/lxn/win"
	"github.com/sirupsen/logrus"
	"os"
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
	createNewDB := false
	logLevelStr := "warn"

	flag.BoolVar(&mustRunPeer, "must-run-peer", true, "ensure peer application")
	flag.BoolVar(&createNewDB, "create-new-db", false, "create new data base")
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

	if createNewDB {
		logrus.Warn("delete data base file because create-new-db flag was set")
		if err := os.Remove(app.DataFileName()); err != nil { // delete data base file
			logrus.WithField("file", app.DataFileName()).Error(err)
		}
	}

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
