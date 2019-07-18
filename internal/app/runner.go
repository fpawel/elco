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

func (_ runner) RunReadAndSaveProductCurrents(field string) {
	runHardware(fmt.Sprintf("Снятие %q", field), func() error {
		return readAndSaveProductsCurrents(field)
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

func (_ runner) RunMainError() {
	runHardware("Снятие основной погрешности", determineMainError)
}

func (_ runner) RunTemperature(workCheck [4]bool) {
	runHardware("Снятие термокомпенсации", func() error {

		ft := func(temperature data.Temperature) func() error {
			return func() error {
				notify.Statusf(log, "%v⁰C: снятие термокомпенсации", temperature)
				return determineTemperature(temperature)
			}
		}
		for i, f := range []func() error{
			ft(20),
			determineMainError,
			ft(-20),
			ft(50),
		} {
			if workCheck[i] {
				if err := f(); err != nil {
					return err
				}
			}
		}
		if err := setupTemperature(20); err != nil {
			return err
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

	log = gohelp.NewLogWithKeys("работа", workName)
	resetLogFunc := func() {
		log = structlog.New()
	}

	go func() {
		defer hardware.WaitGroup.Done()
		defer notify.WorkStoppedf(log, "выполнение окончено: %s", workName)
		defer closeHardware()
		defer resetLogFunc()

		notify.WorkStarted(log, workName)

		if err := portMeasurer.Open(cfg.Cfg.User().ComportMeasurer); err != nil {
			notifyErr(merry.Append(err, "не удалось открыть СОМ порт"))
			return
		}

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

	if portMeasurer.Opened() {
		log.ErrIfFail(portMeasurer.Close)
	}
	if portGas.Opened() {
		log.ErrIfFail(func() error {
			return switchGas(0)
		})
		log.ErrIfFail(portGas.Close)
	}

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
