package app

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/pkg"
	"github.com/fpawel/elco/internal/pkg/ktx500"
	"github.com/powerman/structlog"
	"sync"
	"time"
)

type runner struct{}

const (
	wrOk api.WorkResultTag = iota
	wrCanceled
	wrError
)

func (_ runner) CopyParty(partyID int64) {

	var party data.Party
	if err := data.DB.FindByPrimaryKeyTo(&party, partyID); err != nil {
		panic(err)
	}
	strWhat := fmt.Sprintf("%d %s", party.PartyID, party.CreatedAt.Format("02.01.2006"))

	notifyWnd.WorkStarted(nil, "копирование загрузки: "+strWhat)
	log.Info(strWhat)

	party.PartyID = 0
	party.Note = sql.NullString{"копирование загрузки: " + strWhat, true}
	party.CreatedAt = time.Now()
	if err := data.DB.Save(&party); err != nil {
		panic(err)
	}

	xsProducts, err := data.DB.SelectAllFrom(data.ProductTable, "WHERE party_id = ?", partyID)
	if err != nil {
		panic(err)
	}
	for _, p := range xsProducts {
		p := p.(*data.Product)
		p.Note = sql.NullString{
			String: fmt.Sprintf("копия %d из загрузки %s", p.ProductID, strWhat),
			Valid:  true,
		}
		p.ProductID = 0
		p.PartyID = party.PartyID

		if err = data.DB.Save(p); err != nil {
			panic(err)
		}
	}
	notifyWnd.WorkComplete(nil, api.WorkResult{"копирование загрузки: " + strWhat, wrOk, "успешно"})
	notifyWnd.LastPartyChanged(nil, api.LastParty1())
}

func (_ runner) RunReadAndSaveProductCurrents(dbColumn string) {
	runWork(fmt.Sprintf("Снятие %q", dbColumn), func(x worker) error {
		return x.readSaveForDBColumn(dbColumn)
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
		notifyWnd.ReadFirmware(x.log.Info, data.FirmwareBytes(b).FirmwareInfo(place))
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
		return x.switchGas(n)
	})
}

func (_ runner) RunMain(nku, variation, minus, plus bool) {

	runWork("Снятие", func(x worker) error {
		defer func() {
			_ = x.perform("остановка работы оборудования", func(x worker) error {
				x.ctx = context.Background()
				x.log.ErrIfFail(x.portMeasurer.Close, "main_work_close", "`закрыть СОМ-порт стенда`")
				if x.portGas.Opened() {
					x.log.ErrIfFail(x.switchGasOff, "main_work_close", "`отключить газ`")
					x.log.ErrIfFail(x.portGas.Close, "main_work_close", "`закрыть СОМ-порт пневмоблока`")
				}
				if i, err := ktx500.GetLast(); err == nil {
					if i.TemperatureOn {
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

		if err := x.portMeasurer.Open(); err != nil {
			return err
		}

		setupAndHoldTemperature := func(temperature data.Temperature) error {
			if err := setupTemperature(x, float64(temperature)); err != nil {
				return err
			}
			duration := time.Minute * time.Duration(cfg.Cfg.Gui().HoldTemperatureMinutes)
			return delayf(x, duration, "выдержка T=%v⁰C", temperature)
		}

		if nku || variation {
			if err := setupAndHoldTemperature(20); err != nil {
				return err
			}
		}
		if nku {
			if err := x.readSaveAtTemperature(20); err != nil {
				return err
			}
		}
		if variation {
			if err := x.readSaveForMainError(); err != nil {
				return err
			}
		}
		if minus {
			if err := setupAndHoldTemperature(-20); err != nil {
				return err
			}
			if err := x.readSaveAtTemperature(-20); err != nil {
				return err
			}
		}
		if plus {
			if err := setupAndHoldTemperature(50); err != nil {
				return err
			}
			if err := x.readSaveAtTemperature(50); err != nil {
				return err
			}
		}
		if minus || plus {
			if err := setupTemperature(x, float64(20)); err != nil {
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
				pause(x.ctx.Done(), intSeconds(cfg.Cfg.Dev().ReadBlockPauseSeconds))
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

	worker := newWorker(ctxWork, workName)

	go func() {
		defer func() {
			log.ErrIfFail(worker.portMeasurer.Close)
			log.ErrIfFail(worker.portGas.Close)
			wgWork.Done()
		}()

		notifyWnd.WorkStarted(worker.log.Info, workName)
		err := work(worker)
		if err == nil {
			worker.log.Info("выполнено успешно")
			notifyWnd.WorkComplete(worker.log.Info, api.WorkResult{workName, wrOk, "успешно"})
			return
		}

		kvs := merryKeysValues(err)
		if merry.Is(err, context.Canceled) {
			worker.log.Warn("выполнение прервано", kvs...)
			notifyWnd.WorkComplete(worker.log.Info, api.WorkResult{workName, wrCanceled, "перервано"})
			return
		}
		worker.log.PrintErr(err, append(kvs, "stack", pkg.FormatMerryStacktrace(err))...)
		notifyWnd.WorkComplete(worker.log.Info, api.WorkResult{workName, wrError, err.Error()})
	}()
}
