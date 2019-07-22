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
	msgEndDelay
	msgLastPartyChanged
	msgStartServerApplication
	msgReadFirmware
	msgPanic
	msgWriteConsole
	msgReadPlace
	msgReadBlock
)

func ReadCurrent(log *structlog.Logger, arg api.ReadCurrent) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "ReadCurrent", arg, "MSG", msgReadCurrent)
	}
	go W.NotifyJson(uintptr(msgReadCurrent), arg)
}

func ErrorOccurred(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "ErrorOccurred", arg, "MSG", msgErrorOccurred)
	}
	go W.NotifyStr(uintptr(msgErrorOccurred), arg)
}

func ErrorOccurredf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "ErrorOccurred", fmt.Sprintf(format, a...), "MSG", msgErrorOccurred)
	}
	go W.Notifyf(uintptr(msgErrorOccurred), format, a...)
}
func WorkComplete(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "WorkComplete", arg, "MSG", msgWorkComplete)
	}
	go W.NotifyStr(uintptr(msgWorkComplete), arg)
}

func WorkCompletef(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "WorkComplete", fmt.Sprintf(format, a...), "MSG", msgWorkComplete)
	}
	go W.Notifyf(uintptr(msgWorkComplete), format, a...)
}
func WorkStarted(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "WorkStarted", arg, "MSG", msgWorkStarted)
	}
	go W.NotifyStr(uintptr(msgWorkStarted), arg)
}

func WorkStartedf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "WorkStarted", fmt.Sprintf(format, a...), "MSG", msgWorkStarted)
	}
	go W.Notifyf(uintptr(msgWorkStarted), format, a...)
}
func WorkStopped(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "WorkStopped", arg, "MSG", msgWorkStopped)
	}
	go W.NotifyStr(uintptr(msgWorkStopped), arg)
}

func WorkStoppedf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "WorkStopped", fmt.Sprintf(format, a...), "MSG", msgWorkStopped)
	}
	go W.Notifyf(uintptr(msgWorkStopped), format, a...)
}
func Status(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Status", arg, "MSG", msgStatus)
	}
	go W.NotifyStr(uintptr(msgStatus), arg)
}

func Statusf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Status", fmt.Sprintf(format, a...), "MSG", msgStatus)
	}
	go W.Notifyf(uintptr(msgStatus), format, a...)
}
func Ktx500Info(log *structlog.Logger, arg api.Ktx500Info) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Ktx500Info", arg, "MSG", msgKtx500Info)
	}
	go W.NotifyJson(uintptr(msgKtx500Info), arg)
}

func Ktx500Error(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Ktx500Error", arg, "MSG", msgKtx500Error)
	}
	go W.NotifyStr(uintptr(msgKtx500Error), arg)
}

func Ktx500Errorf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Ktx500Error", fmt.Sprintf(format, a...), "MSG", msgKtx500Error)
	}
	go W.Notifyf(uintptr(msgKtx500Error), format, a...)
}
func Warning(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Warning", arg, "MSG", msgWarning)
	}
	go W.NotifyStr(uintptr(msgWarning), arg)
}

func Warningf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Warning", fmt.Sprintf(format, a...), "MSG", msgWarning)
	}
	go W.Notifyf(uintptr(msgWarning), format, a...)
}
func Delay(log *structlog.Logger, arg api.DelayInfo) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Delay", arg, "MSG", msgDelay)
	}
	go W.NotifyJson(uintptr(msgDelay), arg)
}

func EndDelay(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "EndDelay", arg, "MSG", msgEndDelay)
	}
	go W.NotifyStr(uintptr(msgEndDelay), arg)
}

func EndDelayf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "EndDelay", fmt.Sprintf(format, a...), "MSG", msgEndDelay)
	}
	go W.Notifyf(uintptr(msgEndDelay), format, a...)
}
func LastPartyChanged(log *structlog.Logger, arg data.Party) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "LastPartyChanged", arg, "MSG", msgLastPartyChanged)
	}
	go W.NotifyJson(uintptr(msgLastPartyChanged), arg)
}

func StartServerApplication(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "StartServerApplication", arg, "MSG", msgStartServerApplication)
	}
	go W.NotifyStr(uintptr(msgStartServerApplication), arg)
}

func StartServerApplicationf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "StartServerApplication", fmt.Sprintf(format, a...), "MSG", msgStartServerApplication)
	}
	go W.Notifyf(uintptr(msgStartServerApplication), format, a...)
}
func ReadFirmware(log *structlog.Logger, arg data.FirmwareInfo) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "ReadFirmware", arg, "MSG", msgReadFirmware)
	}
	go W.NotifyJson(uintptr(msgReadFirmware), arg)
}

func Panic(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Panic", arg, "MSG", msgPanic)
	}
	go W.NotifyStr(uintptr(msgPanic), arg)
}

func Panicf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "Panic", fmt.Sprintf(format, a...), "MSG", msgPanic)
	}
	go W.Notifyf(uintptr(msgPanic), format, a...)
}
func WriteConsole(log *structlog.Logger, arg string) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "WriteConsole", arg, "MSG", msgWriteConsole)
	}
	go W.NotifyStr(uintptr(msgWriteConsole), arg)
}

func WriteConsolef(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "WriteConsole", fmt.Sprintf(format, a...), "MSG", msgWriteConsole)
	}
	go W.Notifyf(uintptr(msgWriteConsole), format, a...)
}
func ReadPlace(log *structlog.Logger, arg int) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "ReadPlace", arg, "MSG", msgReadPlace)
	}
	go W.NotifyStr(uintptr(msgReadPlace), fmt.Sprintf("%d", arg))
}

func ReadPlacef(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "ReadPlace", fmt.Sprintf(format, a...), "MSG", msgReadPlace)
	}
	go W.Notifyf(uintptr(msgReadPlace), format, a...)
}
func ReadBlock(log *structlog.Logger, arg int) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "ReadBlock", arg, "MSG", msgReadBlock)
	}
	go W.NotifyStr(uintptr(msgReadBlock), fmt.Sprintf("%d", arg))
}

func ReadBlockf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		log.Debug(elco.PeerWindowClassName, "ReadBlock", fmt.Sprintf(format, a...), "MSG", msgReadBlock)
	}
	go W.Notifyf(uintptr(msgReadBlock), format, a...)
}
