package main

import (
	"flag"
	"github.com/fpawel/elco/internal/app"
	"github.com/fpawel/elco/internal/assets"
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/gohelp/winapp"
	"github.com/lxn/win"
	"github.com/powerman/must"
	"github.com/powerman/structlog"
	"os"
	"path/filepath"
)

func main() {

	createNewDB := flag.Bool("new-db", false, "create new data base")
	hideCon := flag.Bool("hide-con", false, "hide console window")
	skipRunUIApp := flag.Bool("skip-run-ui", false, "skip running ui")
	logLevel := flag.String("log.level", os.Getenv(elco.EnvVarLogLevel), "log `level` (debug|info|warn|err)")

	flag.Parse()

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
		}).SetTimeFormat("15:04:05").
		SetLogLevel(structlog.ParseLevel(*logLevel))

	// Преверяем, не было ли приложение запущено ранее.
	// Если было, выдвигаем окно UI приложения на передний план и завершаем процесс.
	if winapp.IsWindow(winapp.FindWindow(elco.ServerWindowClassName)) {
		hWnd := winapp.FindWindow(elco.PeerWindowClassName)
		win.ShowWindow(hWnd, win.SW_RESTORE)
		win.SetForegroundWindow(hWnd)
		structlog.DefaultLogger.Info("elco.exe already executing")
		return
	}

	must.AbortIf = must.PanicIf

	if *hideCon {
		win.ShowWindow(win.GetConsoleWindow(), win.SW_HIDE)
	}
	must.AbortIf(assets.Ensure())
	must.AbortIf(app.Run(*skipRunUIApp, *createNewDB))

}
