package copydata

import (
	"encoding/json"
	"fmt"
	"github.com/fpawel/elco/internal/pkg"
	"github.com/fpawel/elco/internal/pkg/winapp"
)

type W struct {
	windowClassName, peerWindowClassName string
}

func New(windowClassName, peerWindowClassName string) W {
	return W{
		peerWindowClassName: peerWindowClassName,
		windowClassName:     windowClassName,
	}
}

func (x W) SendMessage(msg uintptr, b []byte) bool {

	hWndPeer := winapp.FindWindow(x.peerWindowClassName)
	hWnd := winapp.FindWindow(x.windowClassName)

	if hWnd == 0 || hWndPeer == 0 {
		return false
	}
	return SendMessage(hWnd, hWndPeer, msg, b) != 0
}

func (x W) Notify(msg uintptr, a ...interface{}) bool {
	return x.NotifyStr(msg, fmt.Sprint(a...))
}

func (x W) NotifyStr(msg uintptr, s string) bool {
	return x.SendMessage(msg, pkg.UTF16FromString(s))
}

func (x W) NotifyJson(msg uintptr, param interface{}) bool {
	b, err := json.Marshal(param)
	if err != nil {
		panic(err)
	}
	return x.Notify(msg, string(b))
}

func (x W) Notifyf(msg uintptr, format string, a ...interface{}) bool {
	return x.NotifyStr(msg, fmt.Sprintf(format, a...))
}
