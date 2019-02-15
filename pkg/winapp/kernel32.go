package winapp

import (
	"syscall"
	"unsafe"
)

var (
	kernel32           = mustLoadLibrary("kernel32.dll")
	outputDebugStringF = mustGetProcAddress(kernel32, "OutputDebugStringW")
	createMutexA       = mustGetProcAddress(kernel32, "CreateMutexA")
	//openMutex = mustGetProcAddress(kernel32,"OpenMutex")
)

func OutputDebugString(str string) int32 {

	p, err := syscall.UTF16PtrFromString(str)
	if err != nil {
		panic(err)
	}
	ret, _, _ := syscall.Syscall(outputDebugStringF, 1,
		uintptr(unsafe.Pointer(p)), 0, 0)
	return int32(ret)
}

func CreateMutex(name string, bInitialOwner uintptr) (uintptr, error) {

	p, err := syscall.UTF16PtrFromString(name)

	ret, _, err := syscall.Syscall(createMutexA, 1,
		0, bInitialOwner, uintptr(unsafe.Pointer(p)))

	switch int(err.(syscall.Errno)) {
	case 0:
		return ret, nil
	default:
		return ret, err
	}
}
