package winapp

import (
	"github.com/fpawel/elco/internal/pkg/must"
	"github.com/lxn/win"
	"syscall"
	"unsafe"
)

var (
	libUser32     = mustLoadLibrary("user32.dll")
	isWindow      = mustGetProcAddress(libUser32, "IsWindow")
	getClassNameW = mustGetProcAddress(libUser32, "GetClassNameW")
)

func IsWindow(hWnd win.HWND) bool {
	ret, _, _ := syscall.Syscall(isWindow, 1,
		uintptr(hWnd),
		0,
		0)

	return ret != 0
}

func FindWindow(className string) win.HWND {
	ptrClassName := must.UTF16PtrFromString(className)
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
