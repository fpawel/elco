package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/ktx500"
	"github.com/fpawel/gohelp"
	"github.com/powerman/structlog"
	"sync"
)

type runner struct{}

func (_ runner) RunReadAndSaveProductCurrents(dbColumn string) {
	runHardware(fmt.Sprintf("Снятие %q", dbColumn), func(log *structlog.Logger) error {
		return readSaveForDBColumn(log, dbColumn)
	})
}

func (_ runner) RunWritePlaceFirmware(place int, bytes []byte) {
	runHardware(fmt.Sprintf("Запись прошивки места %s", data.FormatPlace(place)), func(log *structlog.Logger) error {
		err := writePlaceFirmware(log, place, bytes)
		if err != nil {
			return err
		}
		h := newHelperWriteParty()
		h.bytes[place] = bytes
		h.verifyProductsFirmware(log, []int{place})
		return h.error()
	})
}

func (_ runner) RunReadPlaceFirmware(place int) {
	runHardware(fmt.Sprintf("Считывание места %d", place+1), func(log *structlog.Logger) error {
		b, err := readPlaceFirmware(log, place)
		if err != nil {
			return err
		}
		notify.ReadFirmware(log, data.FirmwareBytes(b).FirmwareInfo(place))
		return nil
	})
	return
}

func (_ runner) RunWritePartyFirmware() {
	runHardware("Прошивка партии", writePartyFirmware)
}

func (_ runner) RunSwitchGas(n int) {
	var what string
	if n == 0 {
		what = "отключить газ"
	} else {
		what = fmt.Sprintf("подать газ %d", n)
	}
	runHardware(what, func(log *structlog.Logger) error {
		return switchGasWithoutWarn(log, n)
	})
}

func (_ runner) RunMain(nku, variation, minus, plus bool) {

	runHardware("Снятие", func(log *structlog.Logger) error {

		if nku || variation {
			if err := setupAndHoldTemperature(log, 20); err != nil {
				return err
			}
		}

		if nku {
			if err := readSaveAtTemperature(log, 20); err != nil {
				return err
			}
		}

		if variation {
			if err := readSaveForMainError(log); err != nil {
				return err
			}
		}

		if minus {
			if err := setupAndHoldTemperature(log, -20); err != nil {
				return err
			}
			if err := readSaveAtTemperature(log, -20); err != nil {
				return err
			}
		}
		if plus {
			if err := setupAndHoldTemperature(log, 50); err != nil {
				return err
			}
			if err := readSaveAtTemperature(log, 50); err != nil {
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

	runHardware("опрос", func(log *structlog.Logger) error {
		for {
			checkedBlocks := data.GetLastPartyCheckedBlocks()
			if len(checkedBlocks) == 0 {
				return merry.New("необходимо выбрать блок для опроса")
			}
			for _, block := range checkedBlocks {

				if _, err := readBlockMeasure(log, block, ctxWork); err != nil {
					return err
				}
				pause(ctxWork.Done(), intSeconds(cfg.Cfg.Predefined().ReadBlockPauseSeconds))
			}
		}
	})
}

type WorkFunc = func(log *structlog.Logger) error

func runHardware(workName string, work WorkFunc) {

	log := gohelp.NewLogWithSuffixKeys("работа", workName)

	cancelWorkFunc()
	wgWork.Wait()
	wgWork = sync.WaitGroup{}
	ctxWork, cancelWorkFunc = context.WithCancel(ctxApp)
	wgWork.Add(1)
	go func() {

		defer func() {
			notify.Status(log, "Остановка работы оборудования")

			log.ErrIfFail(portMeasurer.Close, "close_hardware", "port measurer")
			if portGas.Opened() {

				log.ErrIfFail(func() error {
					return switchGasWithoutWarn(log, 0)
				}, "close_hardware", "switch off gas")

				log.ErrIfFail(portGas.Close, "close_hardware", "port gas")

			}

			if i, err := ktx500.GetLast(); err == nil {
				if i.On {
					log.ErrIfFail(func() error {
						return ktx500.WriteCoolOnOff(false)
					}, "close_hardware", "отключение компрессора")
				}
				if i.CoolOn {
					log.ErrIfFail(func() error {
						return ktx500.WriteOnOff(false)
					}, "выключение термокамеры")
				}
			}

			notify.WorkStoppedf(log, "выполнение окончено: %s", workName)
			wgWork.Done()
		}()

		notify.WorkStarted(log, workName)
		err := work(log)
		if err == nil {
			log.Info("выполнено успешно")
			notify.WorkCompletef(log, "%s: выполнено успешно", workName)
			return
		}
		if merry.Is(err, context.Canceled) {
			log.Warn("выполнение прервано")
			return
		}

		kvs := merryKeysValues(err)
		if merry.Is(err, context.Canceled) {
			log.Warn("выполнение прервано", kvs...)
			return
		}
		notify.ErrorOccurredf(log, err.Error())
		log.PrintErr(err, append(kvs, "stack", merryStacktrace(err))...)
	}()
}
