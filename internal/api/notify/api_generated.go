package notify

import (
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/data/journal"
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
	msgWarning
	msgDelay
	msgLastPartyChanged
	msgComportEntry
	msgStartServerApplication
	msgReadFirmware
	msgNewJournalEntry
)

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

func ComportEntry(w W, arg api.ComportEntry) {
	w.NotifyJson(uintptr(msgComportEntry), arg)
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

func NewJournalEntry(w W, arg journal.EntryInfo) {
	w.NotifyJson(uintptr(msgNewJournalEntry), arg)
}
