package main

import (
	"flag"
	"github.com/fpawel/elco/internal"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/gotools/pkg/rungo"
)

func main() {
	args := flag.String("args", "", "command line arguments to pass")
	flag.Parse()
	rungo.Cmd{
		ExeName: "elco.exe",
		ExeArgs: *args,
		UseGUI:  true,
		NotifyGUI: rungo.NotifyGUI{
			MsgCodeConsole: notify.MsgWriteConsole,
			MsgCodePanic:   notify.MsgPanic,
			WindowClass:    internal.PeerWindowClassName,
		},
	}.Exec()
}
