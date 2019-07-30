package app

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/ktx500"
	"github.com/fpawel/gohelp"
	"math"
	"os"
	"time"
)

func setupTemperature(x worker, destinationTemperature float64) error {
	if os.Getenv("ELCO_DEBUG_NO_HARDWARE") == "true" {
		x.log.Warn("skip because ELCO_DEBUG_NO_HARDWARE set")
		return nil
	}

	return x.performf("перевод термокамеры на Т=%v⁰C", destinationTemperature)(func(x worker) error {
		return performWithWarn(x, func() error {
			if err := ktx500.SetupTemperature(destinationTemperature); err != nil {
				return err
			}
			productsWithSerials := data.GetLastPartyProducts(data.WithProduction)
			if len(productsWithSerials) == 0 {
				return merry.New("не выбрано ни одного прибора")
			}
			for {
				for _, products := range groupProductsByBlocks(productsWithSerials) {
					if _, err := readBlockMeasure(x, products[0].Place/8); err != nil {
						return err
					}
					currentTemperature, err := ktx500.ReadTemperature()
					if err != nil {
						return err
					}
					if math.Abs(currentTemperature-destinationTemperature) < 2 {
						return nil
					}
					notify.Statusf(x.log, "ожидание выхода на Т=%v⁰C: %v⁰C", destinationTemperature, currentTemperature)
				}
			}
		})
	})
}

func setupAndHoldTemperature(x worker, temperature data.Temperature) error {
	err := setupTemperature(x, float64(temperature))
	if err != nil {
		return err
	}
	duration := time.Minute * time.Duration(cfg.Cfg.Gui().HoldTemperatureMinutes)
	return delayf(x, duration, "выдержка T=%v⁰C", temperature)
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
func switchGasWithWarn(x worker, n int) error {
	return performWithWarn(x, func() error {
		return switchGasWithoutWarn(x, n)
	})
}

func switchGasWithoutWarn(x worker, n int) error {

	var s string
	if n == 0 {
		s = "отключить газ"
	} else {
		s = fmt.Sprintf("подать ПГС%d", n)
	}
	return x.perform(s, func(x worker) error {
		req := modbus.Request{
			Addr:     5,
			ProtoCmd: 0x10,
			Data: []byte{
				0, 0x32, 0, 1, 2, 0, 0,
			},
		}
		switch n {
		case 0:
			req.Data[6] = 0
		case 1:
			req.Data[6] = 1
		case 2:
			req.Data[6] = 2
		case 3:
			req.Data[6] = 4
		default:
			return merry.Errorf("wrong gas code: %d", n)
		}
		if os.Getenv("ELCO_DEBUG_NO_HARDWARE") == "true" {
			x.log.Warn("skip because ELCO_DEBUG_NO_HARDWARE==true")
			return nil
		}
		x.log.Info("переключение клапана")
		if _, err := req.GetResponse(x.log, x.ctx, x.portGas, nil); err != nil {
			return err
		}
		req = modbus.Request{
			Addr:     1,
			ProtoCmd: 6,
			Data: []byte{
				0, 4, 0, 0,
			},
		}
		if n > 0 {
			req.Data[2] = 0x14
			req.Data[3] = 0xD5
		}
		x.log.Info("установка расхода")
		if _, err := req.GetResponse(x.log, x.ctx, x.portGas, nil); err != nil {
			return err
		}
		return nil
	})
}

func blowGas(x worker, n int) error {
	if err := switchGasWithWarn(x, n); err != nil {
		return err
	}
	duration := time.Minute * time.Duration(cfg.Cfg.Gui().BlowGasMinutes)
	return delayf(x, duration, "продувка ПГС%d", n)
}
