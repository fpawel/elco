package internal

import (
	"github.com/fpawel/elco/internal/pkg/winapp"
	"github.com/lxn/win"
)

const (
	DelphiWindowClassName = "TElcoMainForm"
	WindowClassName       = "ElcoServerWindow"
)

func HWnd() win.HWND {
	return winapp.FindWindow(WindowClassName)
}

//func HWndDelphi() win.HWND{
//	return winapp.FindWindow(DelphiWindowClassName)
//}

func CloseHWnd() {
	win.PostMessage(HWnd(), win.WM_CLOSE, 0, 0)
	winapp.EnumWindowsWithClassName(func(hWnd win.HWND, winClassName string) {
		if winClassName == DelphiWindowClassName {
			win.PostMessage(hWnd, win.WM_CLOSE, 0, 0)
		}
	})
}
