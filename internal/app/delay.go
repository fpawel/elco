package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/gohelp"
	"github.com/fpawel/gohelp/myfmt"
	"time"
)

func delayf(x worker, duration time.Duration, format string, a ...interface{}) error {
	return delay(x, duration, fmt.Sprintf(format, a...))
}

func delay(x worker, duration time.Duration, name string) error {
	fd := myfmt.FormatDuration
	startTime := time.Now()
	x.log = gohelp.LogPrependSuffixKeys(x.log, "start", startTime.Format("15:04:05"))
	var skipDelay context.CancelFunc
	x.ctx, skipDelay = context.WithTimeout(x.ctx, duration)
	skipDelayFunc = func() {
		skipDelay()
		x.log.Info("задержка прервана", "elapsed", myfmt.FormatDuration(time.Since(startTime)))
	}
	ctxWork := x.ctx
	return x.performf("%s: %s", name, fd(duration))(func(x worker) error {
		x.log.Info("задержка начата")
		defer func() {
			x.log.Debug("задержка окончена", "elapsed", fd(time.Since(startTime)))
			notify.EndDelay(x.log, "")
		}()
		for {
			products := data.GetLastPartyProducts(data.WithProduction)
			if len(products) == 0 {
				return merry.New("фоновый опрос: не выбрано ни одного прибора")
			}
			for _, products := range groupProductsByBlocks(products) {

				block := products[0].Place / 8
				notify.Delay(nil, api.DelayInfo{
					What:           name,
					TotalSeconds:   int(duration.Seconds()),
					ElapsedSeconds: int(time.Since(startTime).Seconds()),
				})
				_, err := readBlockMeasure(x, block)
				if err == nil {
					pause(x.ctx.Done(), intSeconds(cfg.Cfg.Dev().ReadBlockPauseSeconds))
					continue
				}
				if merry.Is(err, context.DeadlineExceeded) {
					return nil // задержка истекла
				}
				if merry.Is(err, context.Canceled) {
					if x.ctx.Err() == context.Canceled {
						return nil // задержка пропущена пользователем
					}
					if ctxWork.Err() == context.Canceled {
						return context.Canceled // прервано пользователем
					}
					return nil
				}
				return err
			}
		}
	})
}
