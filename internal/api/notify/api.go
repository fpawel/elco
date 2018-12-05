package notify

import "github.com/fpawel/goutils/copydata"

type Api struct {
	w *copydata.NotifyWindow
}

func NewApi(w *copydata.NotifyWindow) Api {
	return Api{
		w: w,
	}
}
