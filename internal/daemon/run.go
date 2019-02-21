package daemon

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/data/journal"
	"github.com/fpawel/elco/pkg/errfmt"
	"github.com/fpawel/goutils/intrng"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

func (x *D) RunWritePlaceFirmware(place int, bytes []byte) {
	x.runHardware(false, fmt.Sprintf("Запись прошивки места %s", data.FormatPlace(place)), func() error {
		panic("ups!")
		err := x.writeFirmware(place, bytes)
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
		b, err := x.readFirmware(place)
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
	x.runHardware(false, "Прошивка партии", x.writePartyFirmware)
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
					logrus.WithField("temperature", temperature).Errorf("%v", err)
					return err
				}
			}
		}
		return x.determineNKU2()
	})
}

func (x *D) StopHardware() {
	x.hardware.cancel()
}

func (x *D) SkipDelay() {
	x.hardware.skipDelay()
	logrus.Warn("пользователь прервал задержку")
}

func (x *D) RunReadCurrent(checkPlaces [12]bool) {
	var places, xs []int
	for i, v := range checkPlaces {
		if v {
			places = append(places, i)
			xs = append(xs, i+1)
		}
	}
	x.runHardware(false, "Опрос блоков измерительных: "+intrng.Format(xs), func() error {
		for {
			for _, place := range places {
				if _, err := x.readBlockMeasure(place); err != nil {
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

	notify.WorkStarted(x.w, workName)
	x.hardware.WaitGroup.Add(1)

	x.logFields = logrus.Fields{
		"work": workName,
	}
	var currentWork *journal.Work

	if logWork {
		currentWork = &journal.Work{
			Name:      workName,
			CreatedAt: time.Now(),
		}
		if err := x.dbJournal.Save(currentWork); err != nil {
			logrus.Panicln(err)
		}
	}
	x.muCurrentWork.Lock()
	x.currentWork = currentWork
	x.muCurrentWork.Unlock()

	go func() {

		defer func() {
			notify.WorkStoppedf(x.w, "выполнение окончено: %s", workName)
			if logWork {
				x.muCurrentWork.Lock()
				x.currentWork = nil
				x.muCurrentWork.Unlock()
			}
			x.hardware.WaitGroup.Done()
		}()

		notifyErr := func(what string, err error) {
			logrus.WithFields(errfmt.Values(err)).Errorln(err.Error())
			if merry.Is(err, context.Canceled) {
				return
			}
			notify.ErrorOccurredf(x.w, "%s: %v", workName, errfmt.Format(err))
		}

		if err := x.portMeasurer.Open(x.cfg.User().ComportMeasurer, 115200, 0, x.hardware.ctx); err != nil {
			notifyErr("не удалось открыть СОМ порт", err)
			return
		}

		switch err := work(); err {
		case nil:
			logrus.Info("выполнено успешно")
			notify.WorkCompletef(x.w, "%s: выполнено успешно", workName)
		case context.Canceled:
			logrus.Warn("выполнение прервано")
			notify.WorkCompletef(x.w, "%s: выполнение прервано", workName)
		default:
			notifyErr("выполнено с ошибкой:", err)
		}

		if err := x.closeHardware(); err != nil {
			notifyErr("не удалось остановить работу оборудования по завершении настройки:", err)
		}
		for k := range x.logFields {
			delete(x.logFields, k)
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
