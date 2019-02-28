package main

import (
	"flag"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/assets"
	"github.com/fpawel/elco/internal/daemon"
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/elco/pkg/winapp"
	"github.com/lxn/win"
	"github.com/sirupsen/logrus"
	"log"
)

func main() {

	// Преверяем, не было ли приложение запущено ранее
	if winapp.IsWindow(winapp.FindWindow(elco.ServerWindowClassName)) {
		// Если было, выдвигаем окно приложения на передний план и завершаем процесс
		hWnd := winapp.FindWindow(elco.PeerWindowClassName)
		win.ShowWindow(hWnd, win.SW_RESTORE)
		win.SetForegroundWindow(hWnd)
		fmt.Println("elco.exe already executing")
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
	elco.Logger.SetLevel(logLevel)
	logrus.SetLevel(logLevel)
	logrus.SetFormatter(elco.Logger.Formatter)
	logrus.SetOutput(elco.Logger.Out)
	logrus.SetReportCaller(true)

	if err := assets.Ensure(); err != nil {
		log.Fatal(err)
	}

	if err := daemon.Run(skipRunUIApp, createNewDB); err != nil {
		panic(merry.Details(err))
	}
}
