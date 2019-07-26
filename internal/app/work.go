package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/ktx500"
	"github.com/fpawel/gohelp"
	"github.com/prometheus/common/log"
	"math"
	"os"
	"sort"
	"time"
)

func waitTemperature(x worker, destinationTemperature float64) error {

	productsWithSerials := data.GetLastPartyProducts(data.WithSerials)

	if len(productsWithSerials) == 0 {
		return merry.New("не выбрано ни одного прибора")
	}

	for {
		for _, products := range GroupProductsByBlocks(productsWithSerials) {
			currentTemperature, err := ktx500.ReadTemperature()
			if err != nil {
				return err
			}
			if math.Abs(currentTemperature-destinationTemperature) < 2 {
				return nil
			}

			block := products[0].Place / 8
			if _, err = readBlockMeasure(x, block); err != nil {
				return err
			}
			notify.Statusf(x.log, "установка температуры %v⁰C: %v⁰C", destinationTemperature, currentTemperature)
		}
	}
}

func setupTemperature(x worker, destinationTemperature float64) error {
	if os.Getenv("ELCO_DEBUG_NO_HARDWARE") == "true" {
		log.Warn("skip because ELCO_DEBUG_NO_HARDWARE set")
		return nil
	}

	if err := x.performf("установка %v⁰C, фоновый опрос", destinationTemperature)(func(x worker) error {
		return ktx500.SetupTemperature(destinationTemperature)
	}); err != nil {

		if !merry.Is(err, ktx500.Err) {
			return err
		}
		notify.Warningf(x.log, `Не удалось установить температуру: %v⁰C: %v`, destinationTemperature, err)
		if merry.Is(x.ctx.Err(), context.Canceled) {
			return err
		}
		log.Warn("проигнорирована ошибка связи с термокамерой",
			"setup_temperature", destinationTemperature, "error", err)
		return nil
	}

	return x.performf("установка %v⁰C, фоновый опрос", destinationTemperature)(func(x worker) error {
		return waitTemperature(x, destinationTemperature)
	})
}

func setupAndHoldTemperature(x worker, temperature data.Temperature) error {
	err := setupTemperature(x, float64(temperature))
	if err != nil {
		return err
	}
	duration := time.Minute * time.Duration(cfg.Cfg.User().HoldTemperatureMinutes)
	return delayf(x, duration, "выдержка термокамеры: %v⁰C", temperature)
}

func GroupProductsByBlocks(ps []data.Product) (gs [][]*data.Product) {
	m := make(map[int][]*data.Product)
	for i := range ps {
		p := &ps[i]
		v, _ := m[p.Place/8]
		m[p.Place/8] = append(v, p)
	}
	for _, v := range m {
		gs = append(gs, v)
	}
	sort.Slice(gs, func(i, j int) bool {
		return gs[i][0].Place/8 < gs[j][0].Place
	})
	return
}

func readBlockMeasure(x worker, block int) ([]float64, error) {
	x.log = gohelp.LogPrependSuffixKeys(x.log, "блок", block)
	values, err := modbus.Read3BCDs(x.log, x.ctx, x.portMeasurer, modbus.Addr(block+101), 0, 8)
	if err == nil {
		notify.ReadCurrent(nil, api.ReadCurrent{
			Block:  block,
			Values: values,
		})
		return values, nil
	}
	return nil, merry.Appendf(err, "блок измерения %d", block)
}

func init() {
	merry.RegisterDetail("Запрос", "request")
	merry.RegisterDetail("Ответ", "response")
	merry.RegisterDetail("Длительность ожидания", comm.LogKeyDuration)
	merry.RegisterDetail("Порт", "port")
	merry.RegisterDetail("Прибор", "device")
	merry.RegisterDetail("Блок измерительный", "block")
	merry.RegisterDetail("Длительность ожидания статуса", "status_timeout")
	merry.RegisterDetail("Место", "place")
	merry.RegisterDetail("Код статуса", "status")
	merry.RegisterDetail("Адрес", "addr")

}
