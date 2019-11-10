package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/pkg"
	"github.com/powerman/structlog"
	"strings"
	"time"
)

type worker struct {
	log             *structlog.Logger
	ctx             context.Context
	works           []string
	comport         *comport.Port
	comport2        *comport.Port
	comportGas      *comport.Port
	lastGas         *int
	lastTemperature *float64
}

func newWorker(ctx context.Context, name string) worker {
	c := cfg.Get()
	x := worker{
		log:   pkg.NewLogWithSuffixKeys("work", fmt.Sprintf("`%s`", name)),
		ctx:   ctx,
		works: []string{name},
		comport: comport.NewPort(
			comport.Config{
				Baud:        115200,
				ReadTimeout: time.Millisecond,
				Name:        c.ComportName,
			}),

		comportGas: comport.NewPort(comport.Config{
			Baud:        9600,
			ReadTimeout: time.Millisecond,
			Name:        c.ComportGasName,
		}),
	}
	if c.ComportName == c.ComportName2 {
		x.comport2 = x.comport
	} else {
		x.comport2 = comport.NewPort(
			comport.Config{
				Baud:        115200,
				ReadTimeout: time.Millisecond,
				Name:        c.ComportName2,
			})
	}
	return x
}

func (x worker) Reader2() modbus.ResponseReader {
	return x.comport2.NewResponseReader(x.ctx, cfg.Get().Comport)
}

func (x worker) Reader1() modbus.ResponseReader {
	return x.comport.NewResponseReader(x.ctx, cfg.Get().Comport)
}
func (x worker) ReaderGas() modbus.ResponseReader {
	return x.comportGas.NewResponseReader(x.ctx, cfg.Get().ComportGas)
}

func (x worker) withLogKeys(keyvals ...interface{}) worker {
	x.log = pkg.LogPrependSuffixKeys(x.log, keyvals...)
	return x
}

func (x worker) performf(format string, args ...interface{}) func(func(x worker) error) error {
	return func(work func(x worker) error) error {
		return x.perform(fmt.Sprintf(format, args...), work)
	}
}

func (x worker) perform(name string, work func(x worker) error) error {
	x.log.Info("выполнить: " + name)
	x.works = append(x.works, name)
	x.log = pkg.LogPrependSuffixKeys(x.log, fmt.Sprintf("work%d", len(x.works)), fmt.Sprintf("`%s`", name))
	notify.Status(nil, strings.Join(x.works, ": "))
	if err := work(x); err != nil {
		return merry.Append(err, name)
	}
	x.works = x.works[:len(x.works)-1]
	notify.Status(nil, strings.Join(x.works, ": "))
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

	strErr := strings.Join(strings.Split(err.Error(), ": "), "\n\t -")

	notify.Warning(x.log.Warn, fmt.Sprintf("Не удалось выполнить: %s\n\nПричина: %s", x.works[len(x.works)-1], strErr))
	if merry.Is(x.ctx.Err(), context.Canceled) {
		return err
	}
	x.log.Warn("проигнорирована ошибка: " + err.Error())
	return nil
}
