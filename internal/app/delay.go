package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/gohelp"
	"github.com/fpawel/gohelp/helpstr"
	"github.com/powerman/structlog"
	"time"
)

func delay(log *structlog.Logger, what string, duration time.Duration) error {

	fd := helpstr.FormatDuration

	return merry.Appendf(func() error {
		startTime := time.Now()

		log = gohelp.LogPrependSuffixKeys(log,
			"delay", what,
			"duration", fd(duration),
			"start", startTime.Format("15:04:05"),
		)
		log.Info("Задержка: начало")

		if err := func() error {

			ctx, skipDelay := context.WithTimeout(ctxWork, duration)

			skipDelayFunc = func() {
				skipDelay()
				log.Info("задержка прервана",
					"elapsed",
					fd(time.Since(startTime)))
			}

			defer func() {
				notify.EndDelay(log, "")
			}()
			for {
				products := data.GetLastPartyProducts(data.WithSerials, data.WithProduction)

				if len(products) == 0 {
					return merry.New("фоновый опрос: не выбрано ни одного прибора")
				}
				for _, products := range GroupProductsByBlocks(products) {

					if ctx.Err() != nil {
						return nil
					}

					if ctxWork.Err() != nil {
						return ctxWork.Err()
					}

					block := products[0].Place / 8

					log := gohelp.LogPrependSuffixKeys(log, "block", block)

					notify.Delay(nil, api.DelayInfo{
						What:           what,
						TotalSeconds:   int(duration.Seconds()),
						ElapsedSeconds: int(time.Since(startTime).Seconds()),
					})

					_, err := readBlockMeasure(log, block, ctx)

					if err == nil {
						pause(ctx.Done(), intSeconds(cfg.Cfg.Predefined().ReadBlockPauseSeconds))
						continue
					}

					if ctx.Err() != nil {
						return nil
					}

					if ctxWork.Err() != nil {
						return ctxWork.Err()
					}

					notify.Warningf(log, "фоновый опрос: блок измерения %d: %v", block, err)

					if merry.Is(ctxWork.Err(), context.Canceled) {
						return err
					}
					log.Warn("%s: фоновый опрос: проигнорирована ошибка связи с блоком измерительным %d: %v", block, err)
				}
			}
		}(); err != nil {
			return merry.Appendf(err, "%s: %s: %s",
				what,
				fd(duration),
				fd(time.Since(startTime)))
		}
		log.Info("Задержка: выполнено без ошибок")
		return nil
	}(), what+", "+fd(duration))
}
