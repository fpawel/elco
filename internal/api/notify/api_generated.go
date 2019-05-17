package notify

import (
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/pkg/copydata"
)

type msg int

type W = *copydata.NotifyWindow

const (
	msgReadCurrent msg = iota
	msgErrorOccurred
	msgWorkComplete
	msgWorkStarted
	msgWorkStopped
	msgStatus
	msgKtx500Info
	msgKtx500Error
	msgWarning
	msgDelay
	msgLastPartyChanged
	msgStartServerApplication
	msgReadFirmware
	msgPanic
	msgWriteConsole
)

var msgName = map[msg]string{
	msgReadCurrent:            "ReadCurrent",
	msgErrorOccurred:          "ErrorOccurred",
	msgWorkComplete:           "WorkComplete",
	msgWorkStarted:            "WorkStarted",
	msgWorkStopped:            "WorkStopped",
	msgStatus:                 "Status",
	msgKtx500Info:             "Ktx500Info",
	msgKtx500Error:            "Ktx500Error",
	msgWarning:                "Warning",
	msgDelay:                  "Delay",
	msgLastPartyChanged:       "LastPartyChanged",
	msgStartServerApplication: "StartServerApplication",
	msgReadFirmware:           "ReadFirmware",
	msgPanic:                  "Panic",
	msgWriteConsole:           "WriteConsole",
}

func FormatMsg(msgCode uintptr) string {
	s, _ := msgName[msg(msgCode)]
	return s
}

func ReadCurrent(w W, arg api.ReadCurrent) {
	w.NotifyJson(uintptr(msgReadCurrent), arg)
}

func ErrorOccurred(w W, arg string) {
	w.NotifyStr(uintptr(msgErrorOccurred), arg)
}
func ErrorOccurredf(w W, format string, a ...interface{}) {
	w.Notifyf(uintptr(msgErrorOccurred), format, a...)
}
func WorkComplete(w W, arg string) {
	w.NotifyStr(uintptr(msgWorkComplete), arg)
}
func WorkCompletef(w W, format string, a ...interface{}) {
	w.Notifyf(uintptr(msgWorkComplete), format, a...)
}
func WorkStarted(w W, arg string) {
	w.NotifyStr(uintptr(msgWorkStarted), arg)
}
func WorkStartedf(w W, format string, a ...interface{}) {
	w.Notifyf(uintptr(msgWorkStarted), format, a...)
}
func WorkStopped(w W, arg string) {
	w.NotifyStr(uintptr(msgWorkStopped), arg)
}
func WorkStoppedf(w W, format string, a ...interface{}) {
	w.Notifyf(uintptr(msgWorkStopped), format, a...)
}
func Status(w W, arg string) {
	w.NotifyStr(uintptr(msgStatus), arg)
}
func Statusf(w W, format string, a ...interface{}) {
	w.Notifyf(uintptr(msgStatus), format, a...)
}
func Ktx500Info(w W, arg api.Ktx500Info) {
	w.NotifyJson(uintptr(msgKtx500Info), arg)
}

func Ktx500Error(w W, arg string) {
	w.NotifyStr(uintptr(msgKtx500Error), arg)
}
func Ktx500Errorf(w W, format string, a ...interface{}) {
	w.Notifyf(uintptr(msgKtx500Error), format, a...)
}
func Warning(w W, arg string) {
	w.NotifyStr(uintptr(msgWarning), arg)
}
func Warningf(w W, format string, a ...interface{}) {
	w.Notifyf(uintptr(msgWarning), format, a...)
}
func Delay(w W, arg api.DelayInfo) {
	w.NotifyJson(uintptr(msgDelay), arg)
}

func LastPartyChanged(w W, arg data.Party) {
	w.NotifyJson(uintptr(msgLastPartyChanged), arg)
}

func StartServerApplication(w W, arg string) {
	w.NotifyStr(uintptr(msgStartServerApplication), arg)
}
func StartServerApplicationf(w W, format string, a ...interface{}) {
	w.Notifyf(uintptr(msgStartServerApplication), format, a...)
}
func ReadFirmware(w W, arg data.FirmwareInfo) {
	w.NotifyJson(uintptr(msgReadFirmware), arg)
}

func Panic(w W, arg string) {
	w.NotifyStr(uintptr(msgPanic), arg)
}
func Panicf(w W, format string, a ...interface{}) {
	w.Notifyf(uintptr(msgPanic), format, a...)
}
func WriteConsole(w W, arg string) {
	w.NotifyStr(uintptr(msgWriteConsole), arg)
}
func WriteConsolef(w W, format string, a ...interface{}) {
	w.Notifyf(uintptr(msgWriteConsole), format, a...)
}
