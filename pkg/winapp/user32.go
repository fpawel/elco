package winapp

import (
	"github.com/fpawel/goutils"
	"github.com/lxn/win"
	"syscall"
	"unsafe"
)

var (
	libUser32      = mustLoadLibrary("user32.dll")
	isWindow       = mustGetProcAddress(libUser32, "IsWindow")
	getClassNameW  = mustGetProcAddress(libUser32, "GetClassNameW")
	enableMenuItem = mustGetProcAddress(libUser32, "EnableMenuItem")
	getSystemMenu  = mustGetProcAddress(libUser32, "GetSystemMenu")
)

const (
	MF_DISABLED = 0x00000002
	MF_GRAYED   = 0x00000001
)

//winapp.EnableMenuItem(winapp.GetSystemMenu(x.MainWindow.Handle(), win.FALSE), win.SC_CLOSE,
//	win.MF_BYCOMMAND | winapp.MF_DISABLED | winapp.MF_GRAYED)

func EnableMenuItem(hMenu, uIDEnableItem, uEnable uintptr) uintptr {
	r, _, _ := syscall.Syscall(enableMenuItem, 3, hMenu, uIDEnableItem, uEnable)
	return r
}

func GetSystemMenu(hWnd win.HWND, revert uintptr) uintptr {
	r, _, _ := syscall.Syscall(getSystemMenu, 2, uintptr(hWnd), revert, 0)
	return r
}

func IsWindow(hWnd win.HWND) bool {
	ret, _, _ := syscall.Syscall(isWindow, 1,
		uintptr(hWnd),
		0,
		0)

	return ret != 0
}

func FindWindow(className string) win.HWND {
	ptrClassName := goutils.MustUTF16PtrFromString(className)
	return win.FindWindow(ptrClassName, nil)
}

func GetClassName(hWnd win.HWND) (name string, err error) {
	n := make([]uint16, 256)
	p := &n[0]
	r0, _, e1 := syscall.Syscall(getClassNameW, 3, uintptr(hWnd), uintptr(unsafe.Pointer(p)), uintptr(len(n)))
	if r0 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
		return
	}
	name = syscall.UTF16ToString(n)
	return
}

type EnumWindowsWithClassNameCallBack func(hWnd win.HWND, winClassName string)

func EnumWindowsWithClassName(enumWindowsWithClassNameCallBack EnumWindowsWithClassNameCallBack) {

	f := uintptr(syscall.NewCallback(func(hWnd win.HWND, lParam uintptr) uintptr {
		wndClassName, err := GetClassName(hWnd)
		if err != nil {
			panic(err)
		}
		enumWindowsWithClassNameCallBack(hWnd, wndClassName)
		return 1
	}))

	win.EnumChildWindows(0, f, 1)
	return
}
