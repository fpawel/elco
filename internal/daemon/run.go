package daemon

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/data/journal"
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/goutils/intrng"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

func (x *D) RunWriteFirmware(place int, bytes []byte) {
	x.runHardware(fmt.Sprintf("Запись места %d", place+1), func() error {
		return x.writeFirmware(place, bytes)
	})
}

func (x *D) RunReadFirmware(place int) {
	x.runHardware(fmt.Sprintf("Считывание места %d", place+1), func() error {
		b, err := x.readFirmware(place)
		if err != nil {
			return err
		}
		gases, err := data.ListGases(x.db)
		if err != nil {
			return err
		}
		units, err := data.ListUnits(x.db)
		if err != nil {
			return err
		}
		notify.ReadFirmware(x.w, data.FirmwareBytes(b).FirmwareInfo(units, gases))
		return nil
	})
	return
}

func (x *D) RunWritePartyFirmware() {
	x.runHardware("Прошивка партии", x.writePartyFirmware)
}

func (x *D) RunWriteProductFirmware(place int) {
	what := fmt.Sprintf("Прошивка места %d.%d", place/8+1, place%8+1)
	x.runHardware(what, func() error {
		if p, err := data.GetLastPartyProductAtPlace(x.db, place); err != nil {
			return err
		} else {
			return x.writeProductsFirmware([]*data.Product{&p})
		}
	})
}

func (x *D) RunMainError() {
	x.runHardware("Снятие основной погрешности", x.determineMainError)
}

func (x *D) RunTemperature(workCheck [3]bool) {
	x.runHardware("Снятие термокомпенсации", func() error {
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
	x.runHardware("опрос блоков измерительных: "+intrng.Format(xs), func() error {
		x.port.measurer.SetLogger(nil)
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

func (x *D) runHardware(workName string, work WorkFunc) {

	x.hardware.cancel()
	x.hardware.WaitGroup.Wait()
	x.hardware.WaitGroup = sync.WaitGroup{}
	x.hardware.ctx, x.hardware.cancel = context.WithCancel(x.ctx)

	notify.HardwareStarted(x.w, workName)
	x.hardware.WaitGroup.Add(1)
	x.logFields = logrus.Fields{
		"work": workName,
	}
	x.journalWork = &journal.Work{
		Name:      workName,
		CreatedAt: time.Now(),
	}
	if err := x.dbJournal.Save(x.journalWork); err != nil {
		logrus.Panicln(err)
	}

	go func() {

		defer x.port.measurer.SetLogger(elco.Logger)
		defer x.port.gas.SetLogger(elco.Logger)
		x.port.measurer.SetLogger(elco.Logger)
		x.port.gas.SetLogger(elco.Logger)

		notifyErr := func(what string, err error) {
			logrus.Errorln(what, merry.Details(err))
			if !merry.Is(err, context.Canceled) {
				notify.HardwareErrorf(x.w, "%s: %v", workName, merry.Details(err))
			}
		}

		if err := x.port.measurer.Open(x.cfg.User().ComportMeasurer, 115200, 0, x.hardware.ctx); err != nil {
			notifyErr("не удалось открыть СОМ порт", err)
			return
		}

		switch err := work(); err {
		case nil:
			logrus.Info("выполнено успешно")
		case context.Canceled:
			logrus.Warn("выполнение прервано")
		default:
			notifyErr("выполнено с ошибкой:", err)
		}

		if err := x.closeHardware(); err != nil {
			notifyErr("не удалось остановить работу оборудования по завершении настройки:", err)
		}

		notify.HardwareStoppedf(x.w, "выполнение окончено: %s", workName)
		for k := range x.logFields {
			delete(x.logFields, k)
		}
		x.journalWork = nil
		x.hardware.WaitGroup.Done()
	}()
}

func (x *D) closeHardware() (mulErr error) {

	if x.port.measurer.Opened() {
		if err := x.port.measurer.Close(); err != nil {
			mulErr = multierror.Append(mulErr, merry.WithMessagef(err,
				"закрыть СОМ порт блоков измерения по завершении: %s", err.Error()))
		}
	}
	if x.port.gas.Opened() {

		if err := x.switchGas(0); err != nil {
			mulErr = multierror.Append(mulErr, merry.WithMessagef(err,
				"отключение газового блока по завершении: %s", err.Error()))
		}
		if err := x.port.gas.Close(); err != nil {
			mulErr = multierror.Append(mulErr, merry.WithMessagef(err,
				"закрыть СОМ порт газового блока по завершении: %s", err.Error()))
		}
	}
	return
}

func merryValues(err error, m logrus.Fields) {
	for k, v := range merry.Values(err) {
		var s string
		switch k := k.(type) {
		case string:
			s = k
		default:
			s = fmt.Sprintf("%+v", k)
		}
		if _, f := m[s]; !f {
			m[s] = v
		}
	}
	m["stack"] = merry.Stacktrace(err)
	return
}
