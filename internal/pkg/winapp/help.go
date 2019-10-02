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
