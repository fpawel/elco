package notify

import (
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/gohelp/copydata"
)

// окно для отправки сообщений WM_COPYDATA дельфи-приложению
var W *copydata.NotifyWindow

func InitWindow(sourceWindowClassNameSuffix string) {
	W = copydata.NewNotifyWindow(
		elco.ServerWindowClassName+sourceWindowClassNameSuffix,
		elco.PeerWindowClassName)
}
