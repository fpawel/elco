package copydata

import (
	"encoding/json"
	"fmt"
	"github.com/fpawel/elco/pkg/winapp"
	"github.com/fpawel/goutils"
	"github.com/lxn/win"
)

type NotifyWindow struct {
	hWnd, hWndPeer      win.HWND
	peerWindowClassName string
}

type formatMsgFunc = func(uintptr) string

func NewNotifyWindow(windowClassName, peerWindowClassName string) *NotifyWindow {
	return &NotifyWindow{
		peerWindowClassName: peerWindowClassName,
		hWnd:                winapp.NewWindowWithClassName(windowClassName, win.DefWindowProc),
	}
}

func (x *NotifyWindow) CloseWindowR() bool {
	return win.SendMessage(x.hWnd, win.WM_CLOSE, 0, 0) == 0
}

func (x *NotifyWindow) CloseWindow() {
	x.CloseWindowR()
}

func (x *NotifyWindow) sendMsg(msg uintptr, b []byte) {

	if !winapp.IsWindow(x.hWndPeer) {
		x.hWndPeer = winapp.FindWindow(x.peerWindowClassName)
	}
	if winapp.IsWindow(x.hWndPeer) && SendMessage(x.hWnd, x.hWndPeer, msg, b) == 0 {
		x.hWndPeer = 0
	}
}

func (x *NotifyWindow) Notify(msg uintptr, a ...interface{}) {
	x.NotifyStr(msg, fmt.Sprint(a...))
}

func (x *NotifyWindow) NotifyStr(msg uintptr, s string) {
	x.sendMsg(msg, goutils.UTF16FromString(s))
}

func (x *NotifyWindow) NotifyJson(msg uintptr, param interface{}) {
	b, err := json.Marshal(param)
	if err != nil {
		panic(err)
	}
	x.Notify(msg, string(b))
}

func (x *NotifyWindow) Notifyf(msg uintptr, format string, a ...interface{}) {
	x.NotifyStr(msg, fmt.Sprintf(format, a...))
}
