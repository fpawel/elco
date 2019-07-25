package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/gohelp"
	"github.com/powerman/structlog"
)

func readSaveAtTemperature(log *structlog.Logger, temperature data.Temperature) error {

	log = gohelp.LogPrependSuffixKeys(log, "T⁰C", temperature)

	defer func() {
		log.ErrIfFail(func() error {
			return switchGasWithoutWarn(log, 0)
		}, "выключение", "газ")
	}()
	notify.Statusf(log, "Снятие %v⁰C: начало шкалы", temperature)
	if err := blowGas(log, 1); err != nil {
		return err
	}
	if err := readSaveAtTemperatureScalePt(log, temperature, data.Fon); err != nil {
		return err
	}
	notify.Statusf(log, "Снятие %v⁰C: конец шкалы", temperature)
	if err := blowGas(log, 3); err != nil {
		return err
	}
	if err := readSaveAtTemperatureScalePt(log, temperature, data.Sens); err != nil {
		return err
	}
	notify.Statusf(log, "Снятие %v⁰C: продувка воздухом после снятия конца шкалы", temperature)
	if err := blowGas(log, 1); err != nil {
		return err
	}
	if temperature == 20 {
		notify.Statusf(log, "Снятие %v⁰C: начало шкалы, повторное", temperature)
		if err := readSaveForDBColumn(log, "i13"); err != nil {
			return err
		}
	}
	return nil
}

func readSaveForMainError(log *structlog.Logger) error {
	for i, pt := range data.MainErrorPoints {
		msg := fmt.Sprintf("Снятие основной погрешности: %d, ПГС%d, %s", i+1, pt.Code(), pt.Field())
		notify.Status(log, msg)
		if err := blowGas(log, pt.Code()); err != nil {
			return err
		}
		if err := readSaveForDBColumn(log, pt.Field()); err != nil {
			return merry.Append(err, msg)
		}
	}
	return nil
}

func readSaveAtTemperatureScalePt(log *structlog.Logger, temperature data.Temperature, scale data.ScaleType) error {
	return readSaveForDBColumn(log, data.TemperatureScaleField(temperature, scale))
}

func readSaveForDBColumn(log *structlog.Logger, dbColumn string) error {

	log = gohelp.LogPrependSuffixKeys(log, "db_column", dbColumn)
	log.Info("снятие")

	productsToWork := data.GetLastPartyProducts(data.WithSerials, data.WithProduction)
	if len(productsToWork) == 0 {
		return merry.Errorf("снятие \"%s\": не выбрано ни одного прибора", dbColumn)
	}

	log = gohelp.LogPrependSuffixKeys(log, "products", formatProducts(productsToWork))

	blockProducts := GroupProductsByBlocks(productsToWork)
	for _, products := range blockProducts {
		block := products[0].Place / 8

		values, err := readBlockMeasure(log, block, ctxWork)
		for ; err != nil; values, err = readBlockMeasure(log, block, ctxWork) {
			if merry.Is(err, context.Canceled) {
				return err
			}
			notify.WarningSync(log, fmt.Sprintf("блок измерения %d: %v", block+1, err))
			if ctxWork.Err() == context.Canceled {
				return err
			}
		}

		for _, p := range products {

			n := p.Place % 8

			log := gohelp.LogPrependSuffixKeys(log,
				"product_id", p.ProductID,
				"place", data.FormatPlace(p.Place),
				"value", values[n])

			if err := data.SetProductValue(p.ProductID, dbColumn, values[n]); err != nil {
				return log.Err(merry.Append(err, "не удалось сохранить в базе данных"))
			}
			log.Info("сохраненено в базе данных")
		}
	}
	party := data.GetLastParty(data.WithProducts)
	notify.LastPartyChanged(nil, party)
	return nil

}
