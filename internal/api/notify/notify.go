package notify

import (
	"github.com/fpawel/elco/internal"
	"github.com/fpawel/gotools/pkg/copydata"
)

var w = copydata.WndClass{
	Src:  internal.ServerWindowClassName,
	Dest: internal.PeerWindowClassName,
}
