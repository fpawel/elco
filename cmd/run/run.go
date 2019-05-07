package main

import (
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/elco/pkg/copydata"
	"github.com/fpawel/elco/pkg/winapp"
	"github.com/fpawel/elco/pkg/winapp/supervisor"
	"log"
	"os"
)

func main() {
	exeFileName, err := winapp.CurrentDirOrProfileFileName(".elco", "elco.exe")
	if err != nil {
		log.Fatal(err)
	}
	if panicStr := supervisor.ExecuteProcess(exeFileName, os.Args[1:]...); len(panicStr) > 0 {
		w := copydata.NewNotifyWindow(elco.ServerWindowClassName, elco.PeerWindowClassName)
		notify.HostApplicationPanic(w, panicStr)
		w.CloseWindow()
	}
}
