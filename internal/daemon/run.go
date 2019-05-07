package daemon

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/journal"
	"github.com/hashicorp/go-multierror"
	"github.com/powerman/structlog"
	"sync"
	"time"
)

func (x *D) RunReadAndSaveProductCurrents(field string) {
	x.runHardware(true, fmt.Sprintf("Снятие %q", field), func() error {
		return x.readAndSaveProductsCurrents(field)
	})
}

func (x *D) RunWritePlaceFirmware(place int, bytes []byte) {
	x.runHardware(false, fmt.Sprintf("Запись прошивки места %s", data.FormatPlace(place)), func() error {
		err := x.writePlaceFirmware(place, bytes)
		if err != nil {
			return err
		}
		m := map[int][]byte{place: bytes}
		if err := x.verifyProductsFirmware([]int{place}, m); err != nil {
			return err
		}
		return nil
	})
}

func (x *D) RunReadPlaceFirmware(place int) {
	x.runHardware(false, fmt.Sprintf("Считывание места %d", place+1), func() error {
		b, err := x.readPlaceFirmware(place)
		if err != nil {
			return err
		}
		gases, err := data.ListGases(x.dbProducts)
		if err != nil {
			return err
		}
		units, err := data.ListUnits(x.dbProducts)
		if err != nil {
			return err
		}
		notify.ReadFirmware(x.w, data.FirmwareBytes(b).FirmwareInfo(place, units, gases))
		return nil
	})
	return
}

func (x *D) RunWritePartyFirmware() {
	panic("ups")
	x.runHardware(true, "Прошивка партии", x.writePartyFirmware)
}

func (x *D) RunMainError() {
	x.runHardware(true, "Снятие основной погрешности", x.determineMainError)
}

func (x *D) RunTemperature(workCheck [3]bool) {
	x.runHardware(true, "Снятие термокомпенсации", func() error {
		for i, temperature := range []data.Temperature{20, -20, 50} {
			if workCheck[i] {
				notify.Statusf(x.w, "%v⁰C: снятие термокомпенсации", temperature)
				if err := x.determineTemperature(temperature); err != nil {
					return err
				}
			}
		}
		if err := x.setupTemperature(20); err != nil {
			return err
		}
		return nil
	})
}

func (x *D) StopHardware() {
	x.hardware.cancel()
}

func (x *D) SkipDelay() {
	x.hardware.skipDelay()

}

func (x *D) RunReadCurrent() {

	x.runHardware(false, "опрос", func() error {
		for {
			var checkedBlocks []int
			if err := data.GetCheckedBlocks(x.dbxProducts, &checkedBlocks); err != nil {
				return err
			}
			if len(checkedBlocks) == 0 {
				return merry.New("необходимо выбрать блок для опроса")
			}
			for _, block := range checkedBlocks {

				if _, err := x.readBlockMeasure(x.log, block, x.hardware.ctx); err != nil {
					return err
				}
			}
		}
	})
}

type WorkFunc = func() error

func (x *D) runHardware(logWork bool, workName string, work WorkFunc) {

	x.hardware.cancel()
	x.hardware.WaitGroup.Wait()
	x.hardware.WaitGroup = sync.WaitGroup{}
	x.hardware.ctx, x.hardware.cancel = context.WithCancel(x.ctx)

	x.hardware.WaitGroup.Add(1)

	x.log = comm.NewLogWithKeys("работа", "`"+workName+"`")

	var currentWork *journal.Work

	if logWork {
		currentWork = &journal.Work{
			Name:      workName,
			CreatedAt: time.Now(),
		}
		if err := x.dbJournal.Save(currentWork); err != nil {
			panic(err)
		}
	}
	x.muCurrentWork.Lock()
	x.currentWork = currentWork
	x.muCurrentWork.Unlock()

	go func() {
		notify.WorkStarted(x.w, workName)
		defer func() {
			notify.WorkStoppedf(x.w, "выполнение окончено: %s", workName)
			if logWork {
				x.muCurrentWork.Lock()
				x.currentWork = nil
				x.muCurrentWork.Unlock()
			}
			x.hardware.WaitGroup.Done()
			x.log = structlog.New()
		}()

		notifyErr := func(what string, err error) {
			var kvs []interface{}
			for k, v := range merry.Values(err) {
				if fmt.Sprintf("%v", k) != "stack" {
					kvs = append(kvs, k, v)
				}
			}
			kvs = append(kvs, "stack", merry.Stacktrace(err))
			x.log.PrintErr(err, kvs...)

			if merry.Is(err, context.Canceled) {
				return
			}
			notify.ErrorOccurredf(x.w, "%s: %v", workName, err)
		}

		if err := x.portMeasurer.Open(x.cfg.User().ComportMeasurer); err != nil {
			notifyErr("не удалось открыть СОМ порт", err)
			return
		}

		switch err := work(); err {
		case nil:
			x.log.Info("выполнено успешно")
			notify.WorkCompletef(x.w, "%s: выполнено успешно", workName)
		case context.Canceled:
			x.log.Warn("выполнение прервано")
			notify.WorkCompletef(x.w, "%s: выполнение прервано", workName)
		default:
			notifyErr("выполнено с ошибкой:", err)
		}

		if err := x.closeHardware(); err != nil {
			notifyErr("не удалось остановить работу оборудования по завершении настройки:", err)
		}
	}()
}

func (x *D) closeHardware() (mulErr error) {

	if x.portMeasurer.Opened() {
		if err := x.portMeasurer.Close(); err != nil {
			mulErr = multierror.Append(mulErr, merry.WithMessagef(err,
				"закрыть СОМ порт блоков измерения по завершении: %s", err.Error()))
		}
	}
	if x.portGas.Opened() {

		if err := x.switchGas(0); err != nil {
			mulErr = multierror.Append(mulErr, merry.WithMessagef(err,
				"отключение газового блока по завершении: %s", err.Error()))
		}
		if err := x.portGas.Close(); err != nil {
			mulErr = multierror.Append(mulErr, merry.WithMessagef(err,
				"закрыть СОМ порт газового блока по завершении: %s", err.Error()))
		}
	}
	return
}
