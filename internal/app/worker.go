package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/gohelp"
	"github.com/powerman/structlog"
	"time"
)

type worker struct {
	level                 int
	log                   *structlog.Logger
	ctx                   context.Context
	name                  string
	portMeasurer, portGas *comport.ReadWriter
}

func newWorker(ctx context.Context, name string) worker {
	return worker{
		log:  gohelp.NewLogWithSuffixKeys("work", fmt.Sprintf("`%s`", name)),
		ctx:  ctx,
		name: name,
		portMeasurer: comport.NewReadWriter(func() comport.Config {
			return comport.Config{
				Baud:        115200,
				ReadTimeout: time.Millisecond,
				Name:        cfg.Cfg.User().ComportMeasurer,
			}
		}, func() comm.Config {
			return cfg.Cfg.Predefined().ComportMeasurer
		}),
		portGas: comport.NewReadWriter(func() comport.Config {
			return comport.Config{
				Baud:        9600,
				ReadTimeout: time.Millisecond,
				Name:        cfg.Cfg.User().ComportGas,
			}
		}, func() comm.Config {
			return cfg.Cfg.Predefined().ComportGas
		}),
	}
}

func (x worker) withLogKeys(keyvals ...interface{}) worker {
	x.log = gohelp.LogPrependSuffixKeys(x.log, keyvals...)
	return x
}

func (x worker) performf(format string, args ...interface{}) func(func(x worker) error) error {
	return func(work func(x worker) error) error {
		return x.perform(fmt.Sprintf(format, args...), work)
	}
}

func (x worker) perform(name string, work func(x worker) error) error {
	x.log.Info("выполнить: " + name)
	prevName := x.name
	x.name += ": " + name
	x.level++
	x.log = gohelp.LogPrependSuffixKeys(x.log, fmt.Sprintf("work%d", x.level), fmt.Sprintf("`%s`", name))
	notify.Status(nil, x.name)
	if err := work(x); err != nil {
		return merry.Append(err, name)
	}
	notify.Status(nil, prevName)
	return nil
}

func performWithWarn(x worker, work func() error) error {
	err := work()
	if err == nil {
		return nil
	}
	if merry.Is(x.ctx.Err(), context.Canceled) {
		return err
	}
	notify.Warningf(x.log, "Не удалось выполнить %q: %v.\n\nВыполните вручную либо прервите выполнение.", x.name, err)
	if merry.Is(x.ctx.Err(), context.Canceled) {
		return err
	}
	x.log.Warn("проигнорирована ошибка: ", err.Error())
	return nil
}
