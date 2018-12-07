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
	msgHardwareStarted
	msgHardwareStopped
	msgStatus
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
func HardwareStarted(w W, arg string) {
	w.NotifyStr(uintptr(msgHardwareStarted), arg)
}
func HardwareStartedf(w W, format string, a ...interface{}) {
	w.Notifyf(uintptr(msgHardwareStarted), format, a...)
}
func HardwareStopped(w W, arg string) {
	w.NotifyStr(uintptr(msgHardwareStopped), arg)
}
func HardwareStoppedf(w W, format string, a ...interface{}) {
	w.Notifyf(uintptr(msgHardwareStopped), format, a...)
}
func Status(w W, arg string) {
	w.NotifyStr(uintptr(msgStatus), arg)
}
func Statusf(w W, format string, a ...interface{}) {
	w.Notifyf(uintptr(msgStatus), format, a...)
}
