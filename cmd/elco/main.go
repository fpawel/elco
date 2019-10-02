package main

import (
	"flag"
	"github.com/fpawel/elco/internal"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/app"
	"github.com/fpawel/elco/internal/pkg"
	"github.com/fpawel/elco/internal/pkg/ccolor"
	"github.com/fpawel/elco/internal/pkg/logfile"
	"github.com/lxn/win"
	"github.com/powerman/must"
	"github.com/powerman/structlog"
	"io"
	"os"
	"strings"
)

func main() {
	defaultLogLevelStr := os.Getenv("ELCO_LOG_LEVEL")
	if len(strings.TrimSpace(defaultLogLevelStr)) == 0 {
		defaultLogLevelStr = "info"
	}
	hideCon := flag.Bool("hide-con", false, "hide console window")
	logLevel := flag.String("log.level", defaultLogLevelStr, "log `level` (debug|info|warn|err)")

	flag.Parse()

	if *hideCon {
		win.ShowWindow(win.GetConsoleWindow(), win.SW_HIDE)
	}

	// настрока логгирования
	pkg.InitLog()

	logfileOutput := logfile.NewOutput()
	defer structlog.DefaultLogger.ErrIfFail(logfileOutput.Close)

	structlog.
		DefaultLogger.
		SetLogLevel(structlog.ParseLevel(*logLevel)).
		SetKeysFormat(map[string]string{
			"config": " %+[2]v",
			"работа": " %[1]s=`%[2]s`",
			"error":  " %[1]s=`%[2]s`",
		}).
		SetOutput(io.MultiWriter(ccolor.NewWriter(os.Stderr), guiWriter{}, logfileOutput))

	must.AbortIf = must.PanicIf
	must.AbortIf(app.Run())
}

type guiWriter struct{}

func (x guiWriter) Write(p []byte) (int, error) {
	go notify.New(internal.ServerWindowClassName, internal.PeerWindowClassName).WriteConsole(nil, string(p))
	return len(p), nil
}
