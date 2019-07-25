package notify

import (
	"fmt"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/peer"
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
	msgReadFirmware
	msgPanic
	msgWriteConsole
	msgReadPlace
	msgReadBlock
)

func ReadCurrent(log *structlog.Logger, arg api.ReadCurrent) {
	if log != nil {
		msgReadCurrent.Log(log)(peer.WindowClassName+": ReadCurrent: "+fmt.Sprintf("%+v", arg), "MSG", msgReadCurrent)
	}
	go peer.NotifyJson(uintptr(msgReadCurrent), arg)
}
func ReadCurrentSync(log *structlog.Logger, arg api.ReadCurrent) {
	if log != nil {
		msgReadCurrent.Log(log)(peer.WindowClassName+": ReadCurrent: "+fmt.Sprintf("%+v", arg), "MSG", msgReadCurrent)
	}
	peer.NotifyJson(uintptr(msgReadCurrent), arg)
}

func ErrorOccurred(log *structlog.Logger, arg string) {
	if log != nil {
		msgErrorOccurred.Log(log)(peer.WindowClassName+": ErrorOccurred: "+fmt.Sprintf("%+v", arg), "MSG", msgErrorOccurred)
	}
	go peer.NotifyStr(uintptr(msgErrorOccurred), arg)
}
func ErrorOccurredSync(log *structlog.Logger, arg string) {
	if log != nil {
		msgErrorOccurred.Log(log)(peer.WindowClassName+": ErrorOccurred: "+fmt.Sprintf("%+v", arg), "MSG", msgErrorOccurred)
	}
	peer.NotifyStr(uintptr(msgErrorOccurred), arg)
}

func ErrorOccurredf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgErrorOccurred.Log(log)(peer.WindowClassName+": ErrorOccurred: "+fmt.Sprintf(format, a...), "MSG", msgErrorOccurred)
	}
	go peer.Notifyf(uintptr(msgErrorOccurred), format, a...)
}
func ErrorOccurredSyncf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgErrorOccurred.Log(log)(peer.WindowClassName+": ErrorOccurred: "+fmt.Sprintf(format, a...), "MSG", msgErrorOccurred)
	}
	peer.Notifyf(uintptr(msgErrorOccurred), format, a...)
}
func WorkComplete(log *structlog.Logger, arg string) {
	if log != nil {
		msgWorkComplete.Log(log)(peer.WindowClassName+": WorkComplete: "+fmt.Sprintf("%+v", arg), "MSG", msgWorkComplete)
	}
	go peer.NotifyStr(uintptr(msgWorkComplete), arg)
}
func WorkCompleteSync(log *structlog.Logger, arg string) {
	if log != nil {
		msgWorkComplete.Log(log)(peer.WindowClassName+": WorkComplete: "+fmt.Sprintf("%+v", arg), "MSG", msgWorkComplete)
	}
	peer.NotifyStr(uintptr(msgWorkComplete), arg)
}

func WorkCompletef(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgWorkComplete.Log(log)(peer.WindowClassName+": WorkComplete: "+fmt.Sprintf(format, a...), "MSG", msgWorkComplete)
	}
	go peer.Notifyf(uintptr(msgWorkComplete), format, a...)
}
func WorkCompleteSyncf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgWorkComplete.Log(log)(peer.WindowClassName+": WorkComplete: "+fmt.Sprintf(format, a...), "MSG", msgWorkComplete)
	}
	peer.Notifyf(uintptr(msgWorkComplete), format, a...)
}
func WorkStarted(log *structlog.Logger, arg string) {
	if log != nil {
		msgWorkStarted.Log(log)(peer.WindowClassName+": WorkStarted: "+fmt.Sprintf("%+v", arg), "MSG", msgWorkStarted)
	}
	go peer.NotifyStr(uintptr(msgWorkStarted), arg)
}
func WorkStartedSync(log *structlog.Logger, arg string) {
	if log != nil {
		msgWorkStarted.Log(log)(peer.WindowClassName+": WorkStarted: "+fmt.Sprintf("%+v", arg), "MSG", msgWorkStarted)
	}
	peer.NotifyStr(uintptr(msgWorkStarted), arg)
}

func WorkStartedf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgWorkStarted.Log(log)(peer.WindowClassName+": WorkStarted: "+fmt.Sprintf(format, a...), "MSG", msgWorkStarted)
	}
	go peer.Notifyf(uintptr(msgWorkStarted), format, a...)
}
func WorkStartedSyncf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgWorkStarted.Log(log)(peer.WindowClassName+": WorkStarted: "+fmt.Sprintf(format, a...), "MSG", msgWorkStarted)
	}
	peer.Notifyf(uintptr(msgWorkStarted), format, a...)
}
func WorkStopped(log *structlog.Logger, arg string) {
	if log != nil {
		msgWorkStopped.Log(log)(peer.WindowClassName+": WorkStopped: "+fmt.Sprintf("%+v", arg), "MSG", msgWorkStopped)
	}
	go peer.NotifyStr(uintptr(msgWorkStopped), arg)
}
func WorkStoppedSync(log *structlog.Logger, arg string) {
	if log != nil {
		msgWorkStopped.Log(log)(peer.WindowClassName+": WorkStopped: "+fmt.Sprintf("%+v", arg), "MSG", msgWorkStopped)
	}
	peer.NotifyStr(uintptr(msgWorkStopped), arg)
}

func WorkStoppedf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgWorkStopped.Log(log)(peer.WindowClassName+": WorkStopped: "+fmt.Sprintf(format, a...), "MSG", msgWorkStopped)
	}
	go peer.Notifyf(uintptr(msgWorkStopped), format, a...)
}
func WorkStoppedSyncf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgWorkStopped.Log(log)(peer.WindowClassName+": WorkStopped: "+fmt.Sprintf(format, a...), "MSG", msgWorkStopped)
	}
	peer.Notifyf(uintptr(msgWorkStopped), format, a...)
}
func Status(log *structlog.Logger, arg string) {
	if log != nil {
		msgStatus.Log(log)(peer.WindowClassName+": Status: "+fmt.Sprintf("%+v", arg), "MSG", msgStatus)
	}
	go peer.NotifyStr(uintptr(msgStatus), arg)
}
func StatusSync(log *structlog.Logger, arg string) {
	if log != nil {
		msgStatus.Log(log)(peer.WindowClassName+": Status: "+fmt.Sprintf("%+v", arg), "MSG", msgStatus)
	}
	peer.NotifyStr(uintptr(msgStatus), arg)
}

func Statusf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgStatus.Log(log)(peer.WindowClassName+": Status: "+fmt.Sprintf(format, a...), "MSG", msgStatus)
	}
	go peer.Notifyf(uintptr(msgStatus), format, a...)
}
func StatusSyncf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgStatus.Log(log)(peer.WindowClassName+": Status: "+fmt.Sprintf(format, a...), "MSG", msgStatus)
	}
	peer.Notifyf(uintptr(msgStatus), format, a...)
}
func Ktx500Info(log *structlog.Logger, arg api.Ktx500Info) {
	if log != nil {
		msgKtx500Info.Log(log)(peer.WindowClassName+": Ktx500Info: "+fmt.Sprintf("%+v", arg), "MSG", msgKtx500Info)
	}
	go peer.NotifyJson(uintptr(msgKtx500Info), arg)
}
func Ktx500InfoSync(log *structlog.Logger, arg api.Ktx500Info) {
	if log != nil {
		msgKtx500Info.Log(log)(peer.WindowClassName+": Ktx500Info: "+fmt.Sprintf("%+v", arg), "MSG", msgKtx500Info)
	}
	peer.NotifyJson(uintptr(msgKtx500Info), arg)
}

func Ktx500Error(log *structlog.Logger, arg string) {
	if log != nil {
		msgKtx500Error.Log(log)(peer.WindowClassName+": Ktx500Error: "+fmt.Sprintf("%+v", arg), "MSG", msgKtx500Error)
	}
	go peer.NotifyStr(uintptr(msgKtx500Error), arg)
}
func Ktx500ErrorSync(log *structlog.Logger, arg string) {
	if log != nil {
		msgKtx500Error.Log(log)(peer.WindowClassName+": Ktx500Error: "+fmt.Sprintf("%+v", arg), "MSG", msgKtx500Error)
	}
	peer.NotifyStr(uintptr(msgKtx500Error), arg)
}

func Ktx500Errorf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgKtx500Error.Log(log)(peer.WindowClassName+": Ktx500Error: "+fmt.Sprintf(format, a...), "MSG", msgKtx500Error)
	}
	go peer.Notifyf(uintptr(msgKtx500Error), format, a...)
}
func Ktx500ErrorSyncf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgKtx500Error.Log(log)(peer.WindowClassName+": Ktx500Error: "+fmt.Sprintf(format, a...), "MSG", msgKtx500Error)
	}
	peer.Notifyf(uintptr(msgKtx500Error), format, a...)
}
func Warning(log *structlog.Logger, arg string) {
	if log != nil {
		msgWarning.Log(log)(peer.WindowClassName+": Warning: "+fmt.Sprintf("%+v", arg), "MSG", msgWarning)
	}
	go peer.NotifyStr(uintptr(msgWarning), arg)
}
func WarningSync(log *structlog.Logger, arg string) {
	if log != nil {
		msgWarning.Log(log)(peer.WindowClassName+": Warning: "+fmt.Sprintf("%+v", arg), "MSG", msgWarning)
	}
	peer.NotifyStr(uintptr(msgWarning), arg)
}

func Warningf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgWarning.Log(log)(peer.WindowClassName+": Warning: "+fmt.Sprintf(format, a...), "MSG", msgWarning)
	}
	go peer.Notifyf(uintptr(msgWarning), format, a...)
}
func WarningSyncf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgWarning.Log(log)(peer.WindowClassName+": Warning: "+fmt.Sprintf(format, a...), "MSG", msgWarning)
	}
	peer.Notifyf(uintptr(msgWarning), format, a...)
}
func Delay(log *structlog.Logger, arg api.DelayInfo) {
	if log != nil {
		msgDelay.Log(log)(peer.WindowClassName+": Delay: "+fmt.Sprintf("%+v", arg), "MSG", msgDelay)
	}
	go peer.NotifyJson(uintptr(msgDelay), arg)
}
func DelaySync(log *structlog.Logger, arg api.DelayInfo) {
	if log != nil {
		msgDelay.Log(log)(peer.WindowClassName+": Delay: "+fmt.Sprintf("%+v", arg), "MSG", msgDelay)
	}
	peer.NotifyJson(uintptr(msgDelay), arg)
}

func EndDelay(log *structlog.Logger, arg string) {
	if log != nil {
		msgEndDelay.Log(log)(peer.WindowClassName+": EndDelay: "+fmt.Sprintf("%+v", arg), "MSG", msgEndDelay)
	}
	go peer.NotifyStr(uintptr(msgEndDelay), arg)
}
func EndDelaySync(log *structlog.Logger, arg string) {
	if log != nil {
		msgEndDelay.Log(log)(peer.WindowClassName+": EndDelay: "+fmt.Sprintf("%+v", arg), "MSG", msgEndDelay)
	}
	peer.NotifyStr(uintptr(msgEndDelay), arg)
}

func EndDelayf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgEndDelay.Log(log)(peer.WindowClassName+": EndDelay: "+fmt.Sprintf(format, a...), "MSG", msgEndDelay)
	}
	go peer.Notifyf(uintptr(msgEndDelay), format, a...)
}
func EndDelaySyncf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgEndDelay.Log(log)(peer.WindowClassName+": EndDelay: "+fmt.Sprintf(format, a...), "MSG", msgEndDelay)
	}
	peer.Notifyf(uintptr(msgEndDelay), format, a...)
}
func LastPartyChanged(log *structlog.Logger, arg data.Party) {
	if log != nil {
		msgLastPartyChanged.Log(log)(peer.WindowClassName+": LastPartyChanged: "+fmt.Sprintf("%+v", arg), "MSG", msgLastPartyChanged)
	}
	go peer.NotifyJson(uintptr(msgLastPartyChanged), arg)
}
func LastPartyChangedSync(log *structlog.Logger, arg data.Party) {
	if log != nil {
		msgLastPartyChanged.Log(log)(peer.WindowClassName+": LastPartyChanged: "+fmt.Sprintf("%+v", arg), "MSG", msgLastPartyChanged)
	}
	peer.NotifyJson(uintptr(msgLastPartyChanged), arg)
}

func ReadFirmware(log *structlog.Logger, arg data.FirmwareInfo) {
	if log != nil {
		msgReadFirmware.Log(log)(peer.WindowClassName+": ReadFirmware: "+fmt.Sprintf("%+v", arg), "MSG", msgReadFirmware)
	}
	go peer.NotifyJson(uintptr(msgReadFirmware), arg)
}
func ReadFirmwareSync(log *structlog.Logger, arg data.FirmwareInfo) {
	if log != nil {
		msgReadFirmware.Log(log)(peer.WindowClassName+": ReadFirmware: "+fmt.Sprintf("%+v", arg), "MSG", msgReadFirmware)
	}
	peer.NotifyJson(uintptr(msgReadFirmware), arg)
}

func Panic(log *structlog.Logger, arg string) {
	if log != nil {
		msgPanic.Log(log)(peer.WindowClassName+": Panic: "+fmt.Sprintf("%+v", arg), "MSG", msgPanic)
	}
	go peer.NotifyStr(uintptr(msgPanic), arg)
}
func PanicSync(log *structlog.Logger, arg string) {
	if log != nil {
		msgPanic.Log(log)(peer.WindowClassName+": Panic: "+fmt.Sprintf("%+v", arg), "MSG", msgPanic)
	}
	peer.NotifyStr(uintptr(msgPanic), arg)
}

func Panicf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgPanic.Log(log)(peer.WindowClassName+": Panic: "+fmt.Sprintf(format, a...), "MSG", msgPanic)
	}
	go peer.Notifyf(uintptr(msgPanic), format, a...)
}
func PanicSyncf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgPanic.Log(log)(peer.WindowClassName+": Panic: "+fmt.Sprintf(format, a...), "MSG", msgPanic)
	}
	peer.Notifyf(uintptr(msgPanic), format, a...)
}
func WriteConsole(log *structlog.Logger, arg string) {
	if log != nil {
		msgWriteConsole.Log(log)(peer.WindowClassName+": WriteConsole: "+fmt.Sprintf("%+v", arg), "MSG", msgWriteConsole)
	}
	go peer.NotifyStr(uintptr(msgWriteConsole), arg)
}
func WriteConsoleSync(log *structlog.Logger, arg string) {
	if log != nil {
		msgWriteConsole.Log(log)(peer.WindowClassName+": WriteConsole: "+fmt.Sprintf("%+v", arg), "MSG", msgWriteConsole)
	}
	peer.NotifyStr(uintptr(msgWriteConsole), arg)
}

func WriteConsolef(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgWriteConsole.Log(log)(peer.WindowClassName+": WriteConsole: "+fmt.Sprintf(format, a...), "MSG", msgWriteConsole)
	}
	go peer.Notifyf(uintptr(msgWriteConsole), format, a...)
}
func WriteConsoleSyncf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgWriteConsole.Log(log)(peer.WindowClassName+": WriteConsole: "+fmt.Sprintf(format, a...), "MSG", msgWriteConsole)
	}
	peer.Notifyf(uintptr(msgWriteConsole), format, a...)
}
func ReadPlace(log *structlog.Logger, arg int) {
	if log != nil {
		msgReadPlace.Log(log)(peer.WindowClassName+": ReadPlace: "+fmt.Sprintf("%+v", arg), "MSG", msgReadPlace)
	}
	go peer.NotifyStr(uintptr(msgReadPlace), fmt.Sprintf("%d", arg))
}
func ReadPlaceSync(log *structlog.Logger, arg int) {
	if log != nil {
		msgReadPlace.Log(log)(peer.WindowClassName+": ReadPlace: "+fmt.Sprintf("%+v", arg), "MSG", msgReadPlace)
	}
	peer.NotifyStr(uintptr(msgReadPlace), fmt.Sprintf("%d", arg))
}

func ReadPlacef(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgReadPlace.Log(log)(peer.WindowClassName+": ReadPlace: "+fmt.Sprintf(format, a...), "MSG", msgReadPlace)
	}
	go peer.Notifyf(uintptr(msgReadPlace), format, a...)
}
func ReadPlaceSyncf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgReadPlace.Log(log)(peer.WindowClassName+": ReadPlace: "+fmt.Sprintf(format, a...), "MSG", msgReadPlace)
	}
	peer.Notifyf(uintptr(msgReadPlace), format, a...)
}
func ReadBlock(log *structlog.Logger, arg int) {
	if log != nil {
		msgReadBlock.Log(log)(peer.WindowClassName+": ReadBlock: "+fmt.Sprintf("%+v", arg), "MSG", msgReadBlock)
	}
	go peer.NotifyStr(uintptr(msgReadBlock), fmt.Sprintf("%d", arg))
}
func ReadBlockSync(log *structlog.Logger, arg int) {
	if log != nil {
		msgReadBlock.Log(log)(peer.WindowClassName+": ReadBlock: "+fmt.Sprintf("%+v", arg), "MSG", msgReadBlock)
	}
	peer.NotifyStr(uintptr(msgReadBlock), fmt.Sprintf("%d", arg))
}

func ReadBlockf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgReadBlock.Log(log)(peer.WindowClassName+": ReadBlock: "+fmt.Sprintf(format, a...), "MSG", msgReadBlock)
	}
	go peer.Notifyf(uintptr(msgReadBlock), format, a...)
}
func ReadBlockSyncf(log *structlog.Logger, format string, a ...interface{}) {
	if log != nil {
		msgReadBlock.Log(log)(peer.WindowClassName+": ReadBlock: "+fmt.Sprintf(format, a...), "MSG", msgReadBlock)
	}
	peer.Notifyf(uintptr(msgReadBlock), format, a...)
}
