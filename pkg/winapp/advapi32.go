package winapp

import (
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"syscall"
)

var (
	advApi32                = mustLoadDLL("Advapi32.dll")
	regNotifyChangeKeyValue = mustFindProc(advApi32, "RegNotifyChangeKeyValue")
)

func RegNotifyChangeKeyValue(regKey registry.Key, watchSubtree uintptr, dwNotifyFilter uintptr, hEvent windows.Handle, asynchronous uintptr) error {

	_, _, err := regNotifyChangeKeyValue.Call(uintptr(regKey), watchSubtree, dwNotifyFilter, uintptr(hEvent), asynchronous)

	switch int(err.(syscall.Errno)) {
	case 0:
		return nil
	default:
		return err
	}
}
