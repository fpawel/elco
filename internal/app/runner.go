package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/ktx500"
	"github.com/powerman/structlog"
	"sync"
)

type runner struct{}

func (_ runner) RunReadAndSaveProductCurrents(dbColumn string) {
	runWork(fmt.Sprintf("Снятие %q", dbColumn), func(x worker) error {
		return readSaveForDBColumn(x, dbColumn)
	})
}

func (_ runner) RunWritePlaceFirmware(place int, bytes []byte) {
	runWork(fmt.Sprintf("Запись прошивки места %s", data.FormatPlace(place)), func(x worker) error {
		err := writePlaceFirmware(x, place, bytes)
		if err != nil {
			return err
		}
		h := newHelperWriteParty()
		h.bytes[place] = bytes
		h.verifyProductsFirmware(x, []int{place})
		return h.error()
	})
}

func (_ runner) RunReadPlaceFirmware(place int) {
	runWork(fmt.Sprintf("Считывание места %d", place+1), func(x worker) error {
		b, err := readPlaceFirmware(x, place)
		if err != nil {
			return err
		}
		notify.ReadFirmware(x.log, data.FirmwareBytes(b).FirmwareInfo(place))
		return nil
	})
	return
}

func (_ runner) RunWritePartyFirmware() {
	runWork("Прошивка партии", writePartyFirmware)
}

func (_ runner) RunSwitchGas(n int) {
	var what string
	if n == 0 {
		what = "отключить газ"
	} else {
		what = fmt.Sprintf("подать газ %d", n)
	}
	runWork(what, func(x worker) error {
		return switchGasWithoutWarn(x, n)
	})
}

func (_ runner) RunMain(nku, variation, minus, plus bool) {
	runWork("Снятие", func(x worker) error {
		defer func() {
			_ = x.perform("остановка работы оборудования", func(x worker) error {
				x.ctx = context.Background()
				x.log.ErrIfFail(x.portMeasurer.Close, "main_work_close", "`закрыть СОМ-порт стенда`")
				if x.portGas.Opened() {
					x.log.ErrIfFail(func() error {
						return switchGasWithoutWarn(x, 0)
					}, "main_work_close", "`отключить газ`")
					x.log.ErrIfFail(x.portGas.Close, "main_work_close", "`закрыть СОМ-порт пневмоблока`")
				}
				if i, err := ktx500.GetLast(); err == nil {
					if i.On {
						x.log.ErrIfFail(func() error {
							return ktx500.WriteCoolOnOff(false)
						}, "main_work_close", "`отключение компрессора`")
					}
					if i.CoolOn {
						x.log.ErrIfFail(func() error {
							return ktx500.WriteOnOff(false)
						}, "main_work_close", "выключение термокамеры")
					}
				}
				return nil
			})
		}()
		if nku || variation {
			if err := setupAndHoldTemperature(x, 20); err != nil {
				return err
			}
		}
		if nku {
			if err := readSaveAtTemperature(x, 20); err != nil {
				return err
			}
		}
		if variation {
			if err := readSaveForMainError(x); err != nil {
				return err
			}
		}
		if minus {
			if err := setupAndHoldTemperature(x, -20); err != nil {
				return err
			}
			if err := readSaveAtTemperature(x, -20); err != nil {
				return err
			}
		}
		if plus {
			if err := setupAndHoldTemperature(x, 50); err != nil {
				return err
			}
			if err := readSaveAtTemperature(x, 50); err != nil {
				return err
			}
		}
		return nil
	})
}

func (_ runner) StopHardware() {
	cancelWorkFunc()
}

func (_ runner) SkipDelay() {
	skipDelayFunc()

}

func (_ runner) RunReadCurrent() {

	runWork("опрос", func(x worker) error {
		for {
			checkedBlocks := data.GetLastPartyCheckedBlocks()
			if len(checkedBlocks) == 0 {
				return merry.New("необходимо выбрать блок для опроса")
			}
			for _, block := range checkedBlocks {

				if _, err := readBlockMeasure(x, block); err != nil {
					return err
				}
				pause(x.ctx.Done(), intSeconds(cfg.Cfg.Predefined().ReadBlockPauseSeconds))
			}
		}
	})
}

type WorkFunc = func(log *structlog.Logger) error

func runWork(workName string, work func(x worker) error) {

	cancelWorkFunc()
	wgWork.Wait()
	wgWork = sync.WaitGroup{}
	var ctxWork context.Context
	ctxWork, cancelWorkFunc = context.WithCancel(ctxApp)
	wgWork.Add(1)

	worker := newWorker(ctxWork, fmt.Sprintf("`%s`", workName))

	go func() {
		defer func() {
			notify.WorkStoppedf(worker.log, "выполнение окончено: %s", workName)
			wgWork.Done()
		}()

		notify.WorkStarted(worker.log, workName)
		err := work(worker)
		if err == nil {
			worker.log.Info("выполнено успешно")
			notify.WorkCompletef(worker.log, "%s: выполнено успешно", workName)
			return
		}

		kvs := merryKeysValues(err)
		if merry.Is(err, context.Canceled) {
			worker.log.Warn("выполнение прервано", kvs...)
			return
		}
		notify.ErrorOccurredf(worker.log, err.Error())
		worker.log.PrintErr(err, append(kvs, "stack", merryStacktrace(err))...)

		if merry.Is(err, ktx500.Err) {
			worker.log.Warn("ОШИБКА УПРАВЛЕНИЯ ТЕРМОКАМЕРОЙ")
			return
		}
	}()
}
