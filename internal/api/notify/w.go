package notify

import (
	"github.com/powerman/structlog"
)

var infoMSGs = map[msg]struct{}{
	msgWorkStarted:  {},
	msgWorkComplete: {},
}

func (x msg) Log(log *structlog.Logger) func(interface{}, ...interface{}) {
	if _, f := infoMSGs[x]; f {
		return log.Info
	}
	return log.Debug
}
