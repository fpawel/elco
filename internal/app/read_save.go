package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/gohelp"
)

func readSaveAtTemperature(x worker, temperature data.Temperature) error {
	return x.performf("T=%v⁰C", temperature)(func(x worker) error {
		blowReadSaveScalePt := func(scale data.ScaleType) error {
			s := "в начале шкалы"
			if scale == data.Sens {
				s = "на конце шкалы"
			}
			return x.perform(s, func(x worker) error {
				gas := 1
				if scale == data.Sens {
					gas = 3
				}
				if err := blowGas(x, gas); err != nil {
					return err
				}
				return readSaveForDBColumn(x, data.TemperatureScaleField(temperature, scale))
			})
		}
		defer func() {
			x.log.ErrIfFail(func() error {
				return switchGasWithoutWarn(x, 0)
			}, "выключение", "газ")
		}()

		if err := blowReadSaveScalePt(data.Fon); err != nil {
			return err
		}
		if err := blowReadSaveScalePt(data.Sens); err != nil {
			return err
		}
		if err := x.perform("продувка воздухом после снятия конца шкалы", func(x worker) error {
			return blowGas(x, 1)
		}); err != nil {
			return err
		}
		if temperature == 20 {
			if err := x.perform("начало шкалы, повторное", func(x worker) error {
				return readSaveForDBColumn(x, "i13")
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func readSaveForMainError(x worker) error {
	return x.perform("снятие основной погрешности", func(x worker) error {
		for i, pt := range data.MainErrorPoints {
			err := x.performf("%d, ПГС%d, %s", i+1, pt.Code(), pt.Field())(func(x worker) error {
				if err := blowGas(x, pt.Code()); err != nil {
					return err
				}
				return readSaveForDBColumn(x, pt.Field())
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func readSaveForDBColumn(x worker, dbColumn string) error {
	return x.performf("снятие колоки %q", dbColumn)(func(x worker) error {
		productsToWork := data.GetLastPartyProducts(data.WithSerials, data.WithProduction)
		if len(productsToWork) == 0 {
			return merry.Errorf("снятие \"%s\": не выбрано ни одного прибора", dbColumn)
		}

		x.log = gohelp.LogPrependSuffixKeys(x.log, "products", formatProducts(productsToWork))

		blockProducts := GroupProductsByBlocks(productsToWork)
		for _, products := range blockProducts {
			block := products[0].Place / 8
			values, err := readBlockMeasure(x, block)
			for ; err != nil; values, err = readBlockMeasure(x, block) {
				if merry.Is(err, context.Canceled) {
					return err
				}
				notify.Warning(x.log, fmt.Sprintf("блок измерения %d: %v", block+1, err))
				if x.ctx.Err() == context.Canceled {
					return err
				}
			}
			for _, p := range products {
				n := p.Place % 8
				log := gohelp.LogPrependSuffixKeys(x.log,
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
	})
}
