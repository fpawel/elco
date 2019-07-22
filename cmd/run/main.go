package main

import (
	"flag"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/gorunex/pkg/gorunex"
	"github.com/powerman/must"
	"log"
	"os"
)

func main() {

	log.SetFlags(log.Ltime)
	var exeName, args string
	flag.StringVar(&args, "args", "", "command line arguments to pass")
	flag.StringVar(&exeName, "exe", "elco.exe", "path to elco.exe")
	flag.Parse()
	log.Println("log file:", gorunex.LogFileName())
	notify.InitWindow("_" + os.Args[0])

	defer notify.W.Close()
	must.AbortIf(gorunex.Process(exeName, args, func() {
		notify.Panic(nil, "Произошла ошибка ПО. Подробности в лог-файле "+gorunex.LogFileName())
	}, notifyWriter{}))
}

type notifyWriter struct{}

func (x notifyWriter) Write(b []byte) (int, error) {
	notify.WriteConsole(nil, string(b))
	return len(b), nil
}
