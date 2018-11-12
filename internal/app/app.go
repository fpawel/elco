package app

import "github.com/fpawel/goutils/winapp"

const (
	AppName winapp.AnalitpriborAppName = "elco"
)
const (
	PeerExeName = "elcoui.exe"
)

func DataFileName() string {
	return AppName.DataFileName("elco.sqlite")
}
