package winapp

import (
	"github.com/fpawel/elco/internal/pkg/must"
	"github.com/lxn/walk"
	"github.com/lxn/win"
	"syscall"
)

type WindowProcedure = func(hWnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr

func NewWindowWithClassName(windowClassName string, windowProcedure WindowProcedure) win.HWND {

	wndProc := func(hWnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
		switch msg {
		case win.WM_DESTROY:
			win.PostQuitMessage(0)
		default:
			return windowProcedure(hWnd, msg, wParam, lParam)
		}
		return 0
	}

	walk.MustRegisterWindowClassWithWndProcPtr(
		windowClassName, syscall.NewCallback(wndProc))

	return win.CreateWindowEx(
		0,
		must.UTF16PtrFromString(windowClassName),
		nil,
		0,
		0,
		0,
		0,
		0,
		win.HWND_TOP,
		0,
		win.GetModuleHandle(nil),
		nil)
}
