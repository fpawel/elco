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
	"github.com/fpawel/gohelp/helpstr"
	"time"
)

func delayf(x worker, duration time.Duration, format string, a ...interface{}) error {
	return delay(x, duration, fmt.Sprintf(format, a...))
}

func delay(x worker, duration time.Duration, name string) error {
	fd := helpstr.FormatDuration
	startTime := time.Now()
	x.log = gohelp.LogPrependSuffixKeys(x.log, "start", startTime.Format("15:04:05"))
	var skipDelay context.CancelFunc
	x.ctx, skipDelay = context.WithTimeout(x.ctx, duration)
	skipDelayFunc = func() {
		skipDelay()
		x.log.Info("задержка прервана", "elapsed", helpstr.FormatDuration(time.Since(startTime)))
	}
	ctxWork := x.ctx
	return x.performf("%s: %s: %s", x.name, name, fd(duration))(func(x worker) error {
		x.log.Info("задержка начата")
		defer func() {
			x.log.Debug("задержка окончена", "elapsed", fd(time.Since(startTime)))
			notify.EndDelay(x.log, "")
		}()
		for {
			products := data.GetLastPartyProducts(data.WithSerials, data.WithProduction)
			if len(products) == 0 {
				return merry.New("фоновый опрос: не выбрано ни одного прибора")
			}
			for _, products := range GroupProductsByBlocks(products) {

				block := products[0].Place / 8
				notify.Delay(nil, api.DelayInfo{
					What:           name,
					TotalSeconds:   int(duration.Seconds()),
					ElapsedSeconds: int(time.Since(startTime).Seconds()),
				})

				_, err := readBlockMeasure(x, block)
				if err == nil {
					pause(x.ctx.Done(), intSeconds(cfg.Cfg.Predefined().ReadBlockPauseSeconds))
					continue
				}
				if x.ctx.Err() != nil {
					return nil
				}
				if ctxWork.Err() != nil {
					return x.ctx.Err()
				}
				notify.Warningf(x.log, "фоновый опрос: блок измерения %d: %v", block, err)

				if merry.Is(ctxWork.Err(), context.Canceled) {
					return err
				}
				x.log.Warn("%s: фоновый опрос: проигнорирована ошибка связи с блоком измерительным %d: %v", block, err)
			}
		}
	})
}
