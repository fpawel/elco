package winapp

import (
	"github.com/lxn/win"
	"syscall"
)

func MsgBox(message, title string, style uint32) {
	win.MessageBox(
		0,
		mustUTF16Ptr(message),
		mustUTF16Ptr(title),
		style)
}

func mustUTF16Ptr(s string) *uint16 {
	p, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}
	return p
}
