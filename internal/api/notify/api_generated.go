package notify

import (
	"fmt"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/elco"
	"github.com/powerman/structlog"
)

type msg int

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
	msgReadPlace
)

func ReadCurrent(log *structlog.Logger, arg api.ReadCurrent) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "ReadCurrent", arg, "MSG", msgReadCurrent)
	}
	W.NotifyJson(uintptr(msgReadCurrent), arg)
}

func ErrorOccurred(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "ErrorOccurred", arg, "MSG", msgErrorOccurred)
	}
	W.NotifyStr(uintptr(msgErrorOccurred), arg)
}

func ErrorOccurredf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "ErrorOccurred", fmt.Sprintf(format, a...), "MSG", msgErrorOccurred)
	}
	W.Notifyf(uintptr(msgErrorOccurred), format, a...)
}
func WorkComplete(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "WorkComplete", arg, "MSG", msgWorkComplete)
	}
	W.NotifyStr(uintptr(msgWorkComplete), arg)
}

func WorkCompletef(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "WorkComplete", fmt.Sprintf(format, a...), "MSG", msgWorkComplete)
	}
	W.Notifyf(uintptr(msgWorkComplete), format, a...)
}
func WorkStarted(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "WorkStarted", arg, "MSG", msgWorkStarted)
	}
	W.NotifyStr(uintptr(msgWorkStarted), arg)
}

func WorkStartedf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "WorkStarted", fmt.Sprintf(format, a...), "MSG", msgWorkStarted)
	}
	W.Notifyf(uintptr(msgWorkStarted), format, a...)
}
func WorkStopped(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "WorkStopped", arg, "MSG", msgWorkStopped)
	}
	W.NotifyStr(uintptr(msgWorkStopped), arg)
}

func WorkStoppedf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "WorkStopped", fmt.Sprintf(format, a...), "MSG", msgWorkStopped)
	}
	W.Notifyf(uintptr(msgWorkStopped), format, a...)
}
func Status(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Status", arg, "MSG", msgStatus)
	}
	W.NotifyStr(uintptr(msgStatus), arg)
}

func Statusf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Status", fmt.Sprintf(format, a...), "MSG", msgStatus)
	}
	W.Notifyf(uintptr(msgStatus), format, a...)
}
func Ktx500Info(log *structlog.Logger, arg api.Ktx500Info) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Ktx500Info", arg, "MSG", msgKtx500Info)
	}
	W.NotifyJson(uintptr(msgKtx500Info), arg)
}

func Ktx500Error(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Ktx500Error", arg, "MSG", msgKtx500Error)
	}
	W.NotifyStr(uintptr(msgKtx500Error), arg)
}

func Ktx500Errorf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Ktx500Error", fmt.Sprintf(format, a...), "MSG", msgKtx500Error)
	}
	W.Notifyf(uintptr(msgKtx500Error), format, a...)
}
func Warning(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Warning", arg, "MSG", msgWarning)
	}
	W.NotifyStr(uintptr(msgWarning), arg)
}

func Warningf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Warning", fmt.Sprintf(format, a...), "MSG", msgWarning)
	}
	W.Notifyf(uintptr(msgWarning), format, a...)
}
func Delay(log *structlog.Logger, arg api.DelayInfo) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Delay", arg, "MSG", msgDelay)
	}
	W.NotifyJson(uintptr(msgDelay), arg)
}

func LastPartyChanged(log *structlog.Logger, arg data.Party) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "LastPartyChanged", arg, "MSG", msgLastPartyChanged)
	}
	W.NotifyJson(uintptr(msgLastPartyChanged), arg)
}

func StartServerApplication(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "StartServerApplication", arg, "MSG", msgStartServerApplication)
	}
	W.NotifyStr(uintptr(msgStartServerApplication), arg)
}

func StartServerApplicationf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "StartServerApplication", fmt.Sprintf(format, a...), "MSG", msgStartServerApplication)
	}
	W.Notifyf(uintptr(msgStartServerApplication), format, a...)
}
func ReadFirmware(log *structlog.Logger, arg data.FirmwareInfo) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "ReadFirmware", arg, "MSG", msgReadFirmware)
	}
	W.NotifyJson(uintptr(msgReadFirmware), arg)
}

func Panic(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Panic", arg, "MSG", msgPanic)
	}
	W.NotifyStr(uintptr(msgPanic), arg)
}

func Panicf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Panic", fmt.Sprintf(format, a...), "MSG", msgPanic)
	}
	W.Notifyf(uintptr(msgPanic), format, a...)
}
func WriteConsole(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "WriteConsole", arg, "MSG", msgWriteConsole)
	}
	W.NotifyStr(uintptr(msgWriteConsole), arg)
}

func WriteConsolef(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "WriteConsole", fmt.Sprintf(format, a...), "MSG", msgWriteConsole)
	}
	W.Notifyf(uintptr(msgWriteConsole), format, a...)
}
func ReadPlace(log *structlog.Logger, arg int) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "ReadPlace", arg, "MSG", msgReadPlace)
	}
	W.NotifyStr(uintptr(msgReadPlace), fmt.Sprintf("%d", arg))
}

func ReadPlacef(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "ReadPlace", fmt.Sprintf(format, a...), "MSG", msgReadPlace)
	}
	W.Notifyf(uintptr(msgReadPlace), format, a...)
}
