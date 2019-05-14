package main

import (
	"flag"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/elco/pkg/copydata"
	"github.com/fpawel/gorunex/pkg/gorunex"
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

	w := copydata.NewNotifyWindow(
		os.Args[0]+"_"+elco.ServerWindowClassName,
		elco.PeerWindowClassName, nil, nil)
	defer w.CloseWindow()

	notifier := notifyWriter{w}
	onPanic := func() {
		notify.Panic(w, "Произошла ошибка ПО. Подробности в лог-файле "+gorunex.LogFileName())
	}

	if err := gorunex.Process(exeName, args, onPanic, notifier); err != nil {
		log.Fatal(err)
	}
}

type notifyWriter struct {
	w *copydata.NotifyWindow
}

func (x notifyWriter) Write(b []byte) (int, error) {
	notify.WriteConsole(x.w, string(b))
	return len(b), nil
}
