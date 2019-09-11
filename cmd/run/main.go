package main

import (
	"flag"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/gotools/pkg/rungo"
	"github.com/powerman/must"
	"log"
	"os"
)

func main() {
	notify.InitServerWindow(os.Args[0])
	log.SetFlags(log.Ltime)
	var exeName, args string
	flag.StringVar(&args, "args", "", "command line arguments to pass")
	flag.StringVar(&exeName, "exe", "elco.exe", "path to elco.exe")
	flag.Parse()
	logFileName := rungo.LogFileName()
	log.Println("log file:", logFileName)
	must.AbortIf(rungo.Process(exeName, args, func() {
		notify.Panic(nil, "Произошла ошибка ПО. Подробности в лог-файле "+logFileName)
	}, notifyWriter{}))
}

type notifyWriter struct{}

func (x notifyWriter) Write(b []byte) (int, error) {
	notify.WriteConsole(nil, string(b))
	return len(b), nil
}
