package main

import (
	"flag"
	"github.com/fpawel/elco/internal/app"
	"github.com/fpawel/elco/internal/pkg"
	"github.com/lxn/win"
	"github.com/powerman/must"
	"github.com/powerman/structlog"
	"os"
	"strings"
)

func main() {
	defaultLogLevelStr := os.Getenv("ELCO_LOG_LEVEL")
	if len(strings.TrimSpace(defaultLogLevelStr)) == 0 {
		defaultLogLevelStr = "debug"
	}
	hideCon := flag.Bool("hide-con", false, "hide console window")
	logLevel := flag.String("log.level", defaultLogLevelStr, "log `level` (debug|info|warn|err)")

	flag.Parse()

	if *hideCon {
		win.ShowWindow(win.GetConsoleWindow(), win.SW_HIDE)
	}

	// настрока логгирования
	pkg.InitLog()

	structlog.
		DefaultLogger.
		SetLogLevel(structlog.ParseLevel(*logLevel)).
		SetKeysFormat(map[string]string{
			"config": " %+[2]v",
			"работа": " %[1]s=`%[2]s`",
			"error":  " %[1]s=`%[2]s`",
		})

	must.AbortIf = must.PanicIf
	must.AbortIf(app.Run())
}
