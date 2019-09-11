package main

import (
	"flag"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/gotools/pkg/rungo"
	"github.com/powerman/must"
)

func main() {
	args := flag.String("args", "", "command line arguments to pass")
	flag.Parse()
	notify.Window.Close()

	must.AbortIf(rungo.Cmd{
		ExeName: "elco.exe",
		ExeArgs: *args,
		UseGUI:  true,
		NotifyGUI: rungo.NotifyGUI{
			MsgCodeConsole: notify.MsgWriteConsole,
			MsgCodePanic:   notify.MsgPanic,
			WindowClass:    notify.PeerWindowClassName,
		},
	}.Exec())
}
