package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/ktx500"
	"github.com/fpawel/gohelp"
	"github.com/powerman/structlog"
	"sync"
)

type runner struct{}

func (_ runner) RunReadAndSaveProductCurrents(dbColumn string) {
	runHardware(fmt.Sprintf("Снятие %q", dbColumn), func() error {
		return readAndSaveCurrents{}.forDBColumn(dbColumn)
	})
}

func (_ runner) RunWritePlaceFirmware(place int, bytes []byte) {
	runHardware(fmt.Sprintf("Запись прошивки места %s", data.FormatPlace(place)), func() error {
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
	runHardware(fmt.Sprintf("Считывание места %d", place+1), func() error {
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
	runHardware(what, func() error {
		return doSwitchGas(n)
	})
}

func (_ runner) RunMain(nku, variation, minus, plus bool) {

	readAndSaveCurrents := readAndSaveCurrents{}

	runHardware("Снятие", func() error {

		if nku || variation {
			if err := setupAndHoldTemperature(20); err != nil {
				return err
			}
		}

		if nku {
			if err := readAndSaveCurrents.atTemperature(20); err != nil {
				return err
			}
		}

		if variation {
			if err := readAndSaveCurrents.forMainError(); err != nil {
				return err
			}
		}

		if minus {
			if err := setupAndHoldTemperature(-20); err != nil {
				return err
			}
			if err := readAndSaveCurrents.atTemperature(-20); err != nil {
				return err
			}
		}
		if plus {
			if err := setupAndHoldTemperature(50); err != nil {
				return err
			}
			if err := readAndSaveCurrents.atTemperature(50); err != nil {
				return err
			}
		}

		return nil
	})
}

func (_ runner) StopHardware() {
	hardware.cancelFunc()
}

func (_ runner) SkipDelay() {
	hardware.skipDelayFunc()

}

func (_ runner) RunReadCurrent() {

	runHardware("опрос", func() error {
		for {
			checkedBlocks := data.GetLastPartyCheckedBlocks()
			if len(checkedBlocks) == 0 {
				return merry.New("необходимо выбрать блок для опроса")
			}
			for _, block := range checkedBlocks {

				if _, err := readBlockMeasure(log, block, hardware.ctx); err != nil {
					return err
				}
			}
		}
	})
}

type WorkFunc = func() error

func runHardware(workName string, work WorkFunc) {

	hardware.cancelFunc()
	hardware.WaitGroup.Wait()
	hardware.WaitGroup = sync.WaitGroup{}
	hardware.ctx, hardware.cancelFunc = context.WithCancel(ctxApp)

	hardware.WaitGroup.Add(1)

	log = gohelp.NewLogWithSuffixKeys("работа", workName)
	resetLogFunc := func() {
		log = structlog.New()
	}

	go func() {
		defer hardware.WaitGroup.Done()
		defer notify.WorkStoppedf(log, "выполнение окончено: %s", workName)
		defer closeHardware()
		defer resetLogFunc()

		notify.WorkStarted(log, workName)

		switch err := work(); err {
		case nil:
			log.Info("выполнено успешно")
			notify.WorkCompletef(log, "%s: выполнено успешно", workName)
		case context.Canceled:
			log.Warn("выполнение прервано")
			notify.WorkCompletef(log, "%s: выполнение прервано", workName)
		default:
			notifyErr(err)
		}
	}()
}

func notifyErr(err error) {
	if merry.Is(err, context.Canceled) {
		log.Warn("выполнение прервано")
		return
	}
	var kvs []interface{}
	for k, v := range merry.Values(err) {
		strK := fmt.Sprintf("%v", k)
		if strK != "stack" && strK != "msg" && strK != "message" {
			kvs = append(kvs, k, v)
		}
	}

	if merry.Is(err, context.Canceled) {
		log.Warn("выполнение прервано", kvs...)
		return
	}
	kvs = append(kvs, "stack", merryStacktrace(err))
	log.PrintErr(err, kvs...)

	notify.ErrorOccurredf(log, err.Error())
}

func closeHardware() {
	log.ErrIfFail(portMeasurer.Close)
	if portGas.Opened() {
		log.ErrIfFail(func() error {
			return switchGas(0)
		})
	}
	log.ErrIfFail(portGas.Close)
	if i, err := ktx500.GetLast(); err == nil {
		if i.On {
			log.ErrIfFail(func() error {
				return ktx500.WriteCoolOnOff(false)
			})
		}
		if i.CoolOn {
			log.ErrIfFail(func() error {
				return ktx500.WriteOnOff(false)
			})
		}
	}
}
