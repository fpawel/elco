package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/pkg"
	"time"
)

func delayf(x worker, duration time.Duration, format string, a ...interface{}) error {
	return delay(x, duration, fmt.Sprintf(format, a...))
}

func delay(x worker, duration time.Duration, name string) error {
	fd := pkg.FormatDuration
	startTime := time.Now()
	x.log = pkg.LogPrependSuffixKeys(x.log, "start", startTime.Format("15:04:05"))
	var skipDelay context.CancelFunc
	x.ctx, skipDelay = context.WithTimeout(x.ctx, duration)
	skipDelayFunc = func() {
		skipDelay()
		x.log.Info("задержка прервана", "elapsed", pkg.FormatDuration(time.Since(startTime)))
	}
	return x.performf("%s: %s", name, fd(duration))(func(x worker) error {
		x.log.Info("задержка начата")
		defer func() {
			x.log.Debug("задержка окончена", "elapsed", fd(time.Since(startTime)))
			notify.EndDelay(x.log.Info, "")
		}()
		for {
			products := data.ProductsWithProduction(data.LastPartyID())
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
				if x.ctx.Err() != nil {
					return nil // задержка истекла или пропущена
				}
				_, _ = readBlockMeasure(x, block)
				pause(x.ctx.Done(), cfg.Get().ReadBlockPause)
			}
		}
	})
}
