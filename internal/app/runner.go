package app

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/data/chipmem"
	"github.com/fpawel/elco/internal/pkg"
	"github.com/fpawel/elco/internal/pkg/ktx500"
	"strconv"
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

	notify.WorkStarted(nil, "копирование загрузки: "+strWhat)
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
	notify.WorkComplete(nil, api.WorkResult{"копирование загрузки: " + strWhat, wrOk, "успешно"})
	notify.LastPartyChanged(nil, api.LastParty1())
}

func (_ runner) RunReadAndSaveProductCurrents(dbColumn string, gas int, temperature data.Temperature) {
	runWork(fmt.Sprintf("Снятие %q", dbColumn), func(x worker) error {
		return x.readSaveForDBColumn(dbColumn, gas, temperature)
	})
}

func (_ runner) RunWritePlaceFirmware(placeDevice, placeProduct int, bytes []byte) error {
	what := fmt.Sprintf("Запись прошивки места %s", data.FormatPlace(placeProduct))
	if placeProduct != placeDevice {
		what += fmt.Sprintf(": место в стенде %s", data.FormatPlace(placeDevice))
	}
	f := chipmem.Bytes(bytes).FirmwareInfo()
	serial, err := strconv.ParseInt(f.Serial, 10, 64)
	if err != nil {
		return merry.Append(err, "серийный номер")
	}
	party := data.LastParty()
	if _, err := data.UpdateProductAtPlace(placeProduct, func(p *data.Product) {
		p.Serial = sql.NullInt64{serial, true}
		if party.ProductTypeName != f.ProductType {
			p.ProductTypeName = sql.NullString{f.ProductType, true}
		}
	}); err != nil {
		return err
	}

	runWork(what, func(x worker) error {

		notify.LastPartyChanged(nil, api.LastParty1())

		if err := writePlaceFirmware(x, placeDevice, bytes); err != nil {
			return err
		}
		if _, err := data.UpdateProductAtPlace(placeProduct, func(p *data.Product) {
			p.Firmware = bytes
		}); err != nil {
			return err
		}
		notify.LastPartyChanged(nil, api.LastParty1())
		return nil
	})
	return nil
}

func (_ runner) RunReadPlaceFirmware(place int) {
	runWork(fmt.Sprintf("Считывание места %d", place+1), func(x worker) error {
		b, err := readPlaceFirmware(x, place)
		if err != nil {
			return err
		}
		notify.ReadFirmware(x.log.Info, chipmem.Bytes(b).FirmwareInfo())
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
				x.log.ErrIfFail(x.comport.Close, "main_work_close", "`закрыть СОМ-порт стенда`")
				if x.comportGas.Opened() {
					x.log.ErrIfFail(x.switchGasOff, "main_work_close", "`отключить газ`")
					x.log.ErrIfFail(x.comportGas.Close, "main_work_close", "`закрыть СОМ-порт пневмоблока`")
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

		if err := x.comport.Open(); err != nil {
			return err
		}

		setupAndHoldTemperature := func(temperature data.Temperature) error {
			if err := setupTemperature(x, float64(temperature)); err != nil {
				return err
			}
			duration := time.Minute * time.Duration(cfg.Get().HoldTemperatureMinutes)
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
				pause(x.ctx.Done(), cfg.Get().ReadBlockPause)
			}
		}
	})
}

func (_ runner) NewParty(serials []int64) {
	runWork("создание новой партия ЭХЯ", func(x worker) error {
		partyID := data.CreateNewParty()
		strSql := `INSERT INTO product ( party_id,  place, production, serial) VALUES ` + ""
		firstInsertRecord := true
		for i, serial := range serials {
			if serial <= 0 {
				continue
			}
			s := ","
			if firstInsertRecord {
				firstInsertRecord = false
				s = ""
			}
			strSql += fmt.Sprintf("%s(%d, %d, TRUE, %d)", s, partyID, i, serial)
		}
		_, err := data.DBx.Exec(strSql)
		if err != nil {
			return err
		}
		notify.LastPartyChanged(log.Info, api.LastParty1())
		return nil
	})
}

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
			log.ErrIfFail(worker.comport.Close)
			log.ErrIfFail(worker.comportGas.Close)
			wgWork.Done()
		}()

		notify.WorkStarted(worker.log.Info, workName)
		err := work(worker)
		if err == nil {
			worker.log.Info("выполнено успешно")
			notify.WorkComplete(worker.log.Info, api.WorkResult{workName, wrOk, "успешно"})
			return
		}

		kvs := merryKeysValues(err)
		if merry.Is(err, context.Canceled) {
			worker.log.Warn("выполнение прервано", kvs...)
			notify.WorkComplete(worker.log.Info, api.WorkResult{workName, wrCanceled, "перервано"})
			return
		}
		worker.log.PrintErr(err, append(kvs, "stack", pkg.FormatMerryStacktrace(err))...)
		notify.WorkComplete(worker.log.Info, api.WorkResult{workName, wrError, err.Error()})
	}()
}
