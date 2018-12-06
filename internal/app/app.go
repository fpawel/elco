package app

import "github.com/fpawel/goutils/winapp"

const (
	AppName               winapp.AnalitpriborAppName = "elco"
	PeerExeName                                      = "elcoui.exe"
	PipeName                                         = `\\.\pipe\elco`
	ServerWindowClassName                            = "ElcoServerWindow"
	PeerWindowClassName                              = "TElcoMainForm"
)

func DataFileName() string {
	return AppName.DataFileName("elco.sqlite")
}
