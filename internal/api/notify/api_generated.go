package notify

import (
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/goutils/copydata"
)

type msg int

type W = *copydata.NotifyWindow

const (
	msgReadCurrent msg = iota
	msgHardwareError
)

func ReadCurrent(w W, arg api.ReadCurrent) {
	w.NotifyJson(uintptr(msgReadCurrent), arg)
}

func HardwareError(w W, arg string) {
	w.NotifyStr(uintptr(msgHardwareError), arg)
}
func HardwareErrorf(w W, format string, a ...interface{}) {
	w.Notifyf(uintptr(msgHardwareError), format, a...)
}
