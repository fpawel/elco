package main

import (
	"flag"
	"github.com/fpawel/elco/internal/daemon"
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/goutils/winapp"
	"github.com/lxn/win"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

func main() {

	// Преверяем, не было ли приложение запущено ранее
	if winapp.IsWindow(winapp.FindWindow(elco.ServerWindowClassName)) {
		// Если было, выдвигаем окно приложения на передний план и завершаем процесс
		hWnd := winapp.FindWindow(elco.PeerWindowClassName)
		win.ShowWindow(hWnd, win.SW_RESTORE)
		win.SetForegroundWindow(hWnd)
		return
	}

	createNewDB := false
	hideCon := false
	logLevelStr := "info"
	skipRunUIApp := false

	flag.BoolVar(&createNewDB, "new-db", false, "create new data base")
	flag.BoolVar(&hideCon, "hide-con", false, "hide console window")
	flag.BoolVar(&skipRunUIApp, "skip-run-ui", false, "skip running ui")
	flag.StringVar(&logLevelStr, "log-level", "info", "use log level")

	flag.Parse()

	if hideCon {
		win.ShowWindow(win.GetConsoleWindow(), win.SW_HIDE)
	}

	// Log as JSON instead of the default ASCII formatter.
	//logrus.SetFormatter(&logrus.JSONFormatter{})

	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.SetLevel(logLevel)
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = time.RFC3339Nano
	logrus.SetFormatter(customFormatter)
	//logrus.SetReportCaller(true)

	if createNewDB {
		logrus.Warn("delete data base file because create-new-db flag was set")
		if err := os.Remove(elco.DataFileName()); err != nil { // delete data base file
			logrus.WithField("file", elco.DataFileName()).Error(err)
		}
	}

	daemon.New().Run(skipRunUIApp)
}
