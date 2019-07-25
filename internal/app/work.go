package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/ktx500"
	"github.com/fpawel/gohelp"
	"github.com/powerman/structlog"
	"math"
	"os"
	"sort"
	"time"
)

func setupTemperature(log *structlog.Logger, destinationTemperature float64) error {

	if os.Getenv("ELCO_DEBUG_NO_HARDWARE") == "true" {
		log.Warn("skip because ELCO_DEBUG_NO_HARDWARE set")
		return nil
	}

	if err := ktx500.SetupTemperature(destinationTemperature); err != nil {

		if !merry.Is(err, ktx500.Err) {
			return err
		}
		notify.WarningSyncf(log, `Не удалось установить температуру: %v⁰C: %v`, destinationTemperature, err)
		if merry.Is(ctxWork.Err(), context.Canceled) {
			return err
		}
		log.Warn("проигнорирована ошибка связи с термокамерой",
			"setup_temperature", destinationTemperature, "error", err)
		return nil
	}

	productsWithSerials := data.GetLastPartyProducts(data.WithSerials)

	if len(productsWithSerials) == 0 {
		return merry.New("фоновый опрос: не выбрано ни одного прибора")
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
			if _, err = readBlockMeasure(
				gohelp.LogPrependSuffixKeys(log, "фоновый_опрос",
					fmt.Sprintf("установка температуры %v⁰C", destinationTemperature)),
				block, ctxWork); err != nil {
				return err
			}
			notify.Statusf(log, "установка температуры %v⁰C: %v⁰C", destinationTemperature, currentTemperature)
		}
	}
}

func setupAndHoldTemperature(log *structlog.Logger, temperature data.Temperature) error {

	err := setupTemperature(log, float64(temperature))
	if err != nil {
		return err
	}

	duration := time.Minute * time.Duration(cfg.Cfg.User().HoldTemperatureMinutes)
	return delay(log, fmt.Sprintf("выдержка термокамеры: %v⁰C", temperature), duration)
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

func readBlockMeasure(log *structlog.Logger, block int, ctx context.Context) ([]float64, error) {

	log = gohelp.LogPrependSuffixKeys(log, "блок", block)

	values, err := modbus.Read3BCDs(log, ctx, portMeasurer, modbus.Addr(block+101), 0, 8)

	if err == nil {
		notify.ReadCurrent(nil, api.ReadCurrent{
			Block:  block,
			Values: values,
		})
		return values, nil
	}
	return nil, merry.WithValue(err, "block", block)
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
