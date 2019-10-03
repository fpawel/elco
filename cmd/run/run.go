package main

import (
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/elco/internal"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/gotools/pkg/copydata"
	"github.com/fpawel/gotools/pkg/logfile"
	"github.com/powerman/structlog"
	"os"
	"path/filepath"
)

func main() {
	log := structlog.New()

	structlog.DefaultLogger.
		SetPrefixKeys(
			structlog.KeyApp,
			structlog.KeyPID, structlog.KeyLevel, structlog.KeyUnit, structlog.KeyTime,
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
		})
	modbus.SetLogKeysFormat()

	log.ErrIfFail(func() error {
		return logfile.Exec(
			copydata.NewWriter(notify.MsgWriteConsole, internal.WindowClassName, internal.DelphiWindowClassName),
			"elco.exe")
	})
}
