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
	"github.com/powerman/structlog"
	"log"
	"os"
	"path/filepath"
)

func main() {

	structlog.DefaultLogger.
		SetPrefixKeys(
			structlog.KeyApp, structlog.KeyPID, structlog.KeyLevel, structlog.KeyUnit, structlog.KeyTime,
		).
		SetDefaultKeyvals(
			structlog.KeyApp, filepath.Base(os.Args[0]),
			structlog.KeySource, structlog.Auto,
		).
		SetSuffixKeys(
			structlog.KeyStack,
		).
		SetSuffixKeys(structlog.KeySource).
		SetKeysFormat(map[string]string{
			structlog.KeyTime:   " %[2]s",
			structlog.KeySource: " %6[2]s",
			structlog.KeyUnit:   " %6[2]s",
			"config":            " %+[2]v",
			"запрос":            " %[1]s=`% [2]X`",
			"ответ":             " %[1]s=`% [2]X`",
			"работа":            " %[1]s=`%[2]s`",
		}).SetTimeFormat("15:04:05")

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
	skipRunUIApp := false

	flag.BoolVar(&createNewDB, "new-db", false, "create new data base")
	flag.BoolVar(&hideCon, "hide-con", false, "hide console window")
	flag.BoolVar(&skipRunUIApp, "skip-run-ui", false, "skip running ui")

	flag.Parse()

	if hideCon {
		win.ShowWindow(win.GetConsoleWindow(), win.SW_HIDE)
	}

	if err := assets.Ensure(); err != nil {
		log.Fatal(err)
	}

	if err := daemon.Run(skipRunUIApp, createNewDB); err != nil {
		panic(merry.Details(err))
	}
}
