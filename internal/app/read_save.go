package app

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/pkg"
	"time"
)

func (x worker) SwitchGasOffInEnd() {
	if !x.portGas.Opened() {
		return
	}
	_ = x.perform("отключить газ по завершении", func(x worker) error {
		x.log.ErrIfFail(x.switchGasOff)
		return nil
	})
}

func (x worker) readSaveAtTemperature(temperature data.Temperature) error {
	return x.performf("снятие при T=%v⁰C", temperature)(func(x worker) error {

		defer x.SwitchGasOffInEnd()

		blowReadSaveScalePt := func(scale data.ScaleType) error {
			s := "снятие в начале шкалы"
			gas := 1
			if scale == data.Sens {
				s = "снятие в конце шкалы"
				gas = 3
				if cfg.Cfg.Gui().EndScaleGas2 {
					gas = 2
				}
			}
			if err := x.perform(s, func(x worker) error { return blowGas(x, gas) }); err != nil {
				return err
			}
			return x.perform(s, func(x worker) error {
				return x.readSaveForDBColumn(
					data.TemperatureScaleField(temperature, scale),
					gas,
					temperature)
			})
		}
		if err := blowReadSaveScalePt(data.Fon); err != nil {
			return err
		}
		if err := blowReadSaveScalePt(data.Sens); err != nil {
			return err
		}
		if err := x.perform("продувка воздухом после снятия конца шкалы",
			func(x worker) error {
				return blowGas(x, 1)
			},
		); err != nil {
			return err
		}
		if temperature == 20 {
			if err := x.perform("начало шкалы, повторное", func(x worker) error {
				return x.readSaveForDBColumn("i13", 1, 20)
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (x worker) readSaveForMainError() error {
	return x.perform("снятие основной погрешности", func(x worker) error {

		defer x.SwitchGasOffInEnd()

		for i, pt := range data.MainErrorPoints {
			err := x.performf("%d, ПГС%d, %s", i+1, pt.Code(), pt.Field())(func(x worker) error {
				if err := blowGas(x, pt.Code()); err != nil {
					return err
				}
				return x.readSaveForDBColumn(pt.Field(), pt.Code(), 20)
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (x worker) readSaveForDBColumn(dbColumn string, gas int, temperature data.Temperature) error {
	return x.performf("снятие %q газ=%d T=%v", dbColumn, gas, temperature)(func(x worker) error {
		productsToWork := data.ProductsWithProduction(data.LastPartyID())
		if len(productsToWork) == 0 {
			return merry.Errorf("снятие \"%s\": не выбрано ни одного прибора", dbColumn)
		}

		x.log = pkg.LogPrependSuffixKeys(x.log, "products", formatProducts(productsToWork))

		blockProducts := groupProductsByBlocks(productsToWork)
		for _, products := range blockProducts {
			block := products[0].Place / 8
			var values []float64

			if err := x.performf("снятие токов блока %d для сохранения", block)(func(x worker) error {
				var err error
				values, err = readBlockMeasure(x, block)
				return err
			}); err != nil {
				return err
			}

			for _, p := range products {
				n := p.Place % 8
				log := pkg.LogPrependSuffixKeys(x.log,
					"product_id", p.ProductID,
					"place", data.FormatPlace(p.Place),
					"gas", gas,
					"temperature", temperature,
					"value", values[n])
				if err := data.SetProductValue(p.ProductID, dbColumn, values[n]); err != nil {
					log.Panic(merry.Append(err, "не удалось сохранить в базе данных"))
				}
				if err := data.DB.Save(&data.ProductCurrent{
					StoredAt:     time.Now(),
					ProductID:    p.ProductID,
					Temperature:  temperature,
					Gas:          gas,
					CurrentValue: values[n],
					Note:         dbColumn,
				}); err != nil {
					log.Panic(merry.Append(err, "не удалось сохранить в базе данных"))
				}
				log.Info("сохраненено в базе данных")
			}
		}
		notify.LastPartyChanged(nil, api.LastParty1())
		return nil
	})
}

func readSaveForLastTemperatureGas(x worker) error {
	return x.perform("снятие для теккущей температуры и газа", func(x worker) error {
		if x.lastTemperature == nil {
			return merry.New("температура не установлена")
		}
		if x.lastGas == nil {
			return merry.New("газ не установлен")
		}

		productsToWork := data.ProductsWithProduction(data.LastPartyID())
		if len(productsToWork) == 0 {
			return merry.New("снятие для теккущей температуры и газа: не выбрано ни одного прибора")
		}

		temperature := data.Temperature(*x.lastTemperature)
		gas := *x.lastGas

		x.log = pkg.LogPrependSuffixKeys(x.log, "products", formatProducts(productsToWork))

		blockProducts := groupProductsByBlocks(productsToWork)
		for _, products := range blockProducts {
			block := products[0].Place / 8
			var values []float64

			if err := x.performf("снятие токов блока %d для сохранения", block)(func(x worker) error {
				var err error
				values, err = readBlockMeasure(x, block)
				return err
			}); err != nil {
				return err
			}

			for _, p := range products {
				n := p.Place % 8
				log := pkg.LogPrependSuffixKeys(x.log,
					"product_id", p.ProductID,
					"place", data.FormatPlace(p.Place),
					"gas", gas,
					"temperature", temperature,
					"value", values[n])

				if err := data.DB.Save(&data.ProductCurrent{
					StoredAt:     time.Now(),
					ProductID:    p.ProductID,
					Temperature:  temperature,
					Gas:          gas,
					CurrentValue: values[n],
					Note:         "сценарий",
				}); err != nil {
					log.Panic(merry.Append(err, "не удалось сохранить в базе данных"))
				}
				log.Info("сохраненено в базе данных")
			}
		}
		return nil
	})
}
