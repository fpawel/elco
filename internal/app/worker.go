package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/pkg"
	"github.com/powerman/structlog"
	"strings"
	"time"
)

type worker struct {
	log          *structlog.Logger
	ctx          context.Context
	works        []string
	portMeasurer *comport.Port
	portGas      *comport.Port
}

func newWorker(ctx context.Context, name string) worker {
	return worker{
		log:   pkg.NewLogWithSuffixKeys("work", fmt.Sprintf("`%s`", name)),
		ctx:   ctx,
		works: []string{name},
		portMeasurer: comport.NewPort(func() comport.Config {
			return comport.Config{
				Baud:        115200,
				ReadTimeout: time.Millisecond,
				Name:        cfg.Cfg.Gui().ComportMeasurer,
			}
		}),
		portGas: comport.NewPort(func() comport.Config {
			return comport.Config{
				Baud:        9600,
				ReadTimeout: time.Millisecond,
				Name:        cfg.Cfg.Gui().ComportGas,
			}
		}),
	}
}

func (x worker) ReaderMeasurer() modbus.ResponseReader {
	return x.portMeasurer.NewResponseReader(x.ctx, cfg.Cfg.Dev().ComportMeasurer)
}
func (x worker) ReaderGas() modbus.ResponseReader {
	return x.portGas.NewResponseReader(x.ctx, cfg.Cfg.Dev().ComportGas)
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
	notifyWnd.Status(nil, strings.Join(x.works, ": "))
	if err := work(x); err != nil {
		return merry.Append(err, name)
	}
	x.works = x.works[:len(x.works)-1]
	notifyWnd.Status(nil, strings.Join(x.works, ": "))
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

	notifyWnd.Warning(x.log.Warn, fmt.Sprintf("Не удалось выполнить: %s\n\nПричина: %s", x.works[len(x.works)-1], strErr))
	if merry.Is(x.ctx.Err(), context.Canceled) {
		return err
	}
	x.log.Warn("проигнорирована ошибка: " + err.Error())
	return nil
}
