package winapp

import (
	"log"
	"syscall"
)

func mustGetProcAddress(lib uintptr, name string) uintptr {
	addr, err := syscall.GetProcAddress(syscall.Handle(lib), name)
	if err != nil {
		log.Panicln("get procedure address:", name, ":", err)
	}

	return uintptr(addr)
}

func mustLoadLibrary(name string) uintptr {
	lib, err := syscall.LoadLibrary(name)
	if err != nil {
		log.Panicln("load library:", name, ":", err)
	}
	return uintptr(lib)
}

func mustLoadDLL(name string) *syscall.DLL {
	dll, err := syscall.LoadDLL("Advapi32.dll")
	if err != nil {
		log.Panicln("load dll:", name, ":", err)
	}
	return dll
}

func mustFindProc(dll *syscall.DLL, name string) *syscall.Proc {
	proc, err := dll.FindProc(name)
	if err != nil {
		log.Panicln("find procedure address:", name, ":", err)
	}
	return proc
}
