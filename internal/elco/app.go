package elco

import "github.com/fpawel/elco/pkg/winapp"

const (
	AppName               winapp.AnalitpriborAppName = "elco"
	PipeName                                         = `\\.\pipe\elco`
	ServerWindowClassName                            = "ElcoServerWindow"
	PeerWindowClassName                              = "TElcoMainForm"
)

func DataFileName() string {
	return AppName.DataFileName("elco.sqlite")
}
