package copydata

import (
	"github.com/lxn/win"
	"reflect"
	"unsafe"
)

type CopyData struct {
	DwData uintptr
	CbData uint32
	LpData uintptr
}

func GetData(ptr unsafe.Pointer) (uintptr, []byte) {
	cd := (*CopyData)(ptr)
	p := PtrSliceFrom(unsafe.Pointer(cd.LpData), int(cd.CbData))
	return cd.DwData, *(*[]byte)(p)
}

func PtrSliceFrom(p unsafe.Pointer, s int) unsafe.Pointer {
	return unsafe.Pointer(&reflect.SliceHeader{Data: uintptr(p), Len: s, Cap: s})
}

func SendMessage(hWndSource, hWndTarget win.HWND, wParam uintptr, b []byte) uintptr {
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&b))
	cd := CopyData{
		CbData: uint32(header.Len),
		LpData: header.Data,
		DwData: uintptr(hWndSource),
	}
	return win.SendMessage(hWndTarget, win.WM_COPYDATA, wParam, uintptr(unsafe.Pointer(&cd)))
}
