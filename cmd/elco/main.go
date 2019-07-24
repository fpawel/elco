package main

import (
	"flag"
	"github.com/fpawel/elco/internal/app"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/peer"
	"github.com/fpawel/gohelp"
	"github.com/lxn/win"
	"github.com/powerman/must"
	"github.com/powerman/structlog"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	defaultLogLevelStr := gohelp.GetEnvWithLog("ELCO_LOG_LEVEL")

	if len(strings.TrimSpace(defaultLogLevelStr)) == 0 {
		defaultLogLevelStr = "info"
	}

	hideCon := flag.Bool("hide-con", false, "hide console window")
	logLevel := flag.String("log.level", defaultLogLevelStr, "log `level` (debug|info|warn|err)")

	flag.Parse()

	if *hideCon {
		win.ShowWindow(win.GetConsoleWindow(), win.SW_HIDE)
	}

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

	must.AbortIf = must.PanicIf

	log := structlog.New()

	cfg.OpenConfig()
	data.Open()

	must.AbortIf(app.Run())
	peer.Close()
	log.ErrIfFail(data.Close, "defer", "close products db")
	log.ErrIfFail(cfg.Cfg.Save, "defer", "save config")

}
