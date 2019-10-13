package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/pkg"
	"github.com/fpawel/elco/internal/pkg/ktx500"
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
			productList := data.ProductsWithProduction(data.LastPartyID())
			if len(productList) == 0 {
				return merry.New("не выбрано ни одного прибора")
			}
			for {
				if x.ctx.Err() != nil {
					return x.ctx.Err()
				}
				for _, products := range groupProductsByBlocks(productList) {
					if _, err := readBlockMeasure(x, products[0].Place/8); merry.Is(err, context.Canceled) {
						return err
					}
					currentTemperature, err := ktx500.ReadTemperature()
					if err != nil {
						return err
					}
					if math.Abs(currentTemperature-destinationTemperature) < 2 {
						x.lastTemperature = &destinationTemperature
						return nil
					}
					notify.Status(x.log.Debug, "ожидание выхода на температуру",
						"destination.Т⁰C", destinationTemperature,
						"current.Т⁰C", currentTemperature)
				}
			}
		})
	})
}

func readBlockMeasure(x worker, block int) ([]float64, error) {
	x.log = pkg.LogPrependSuffixKeys(x.log, "блок", block)
	values, err := modbus.Read3BCDs(x.log, x.ReaderMeasurer(), modbus.Addr(block+101), 0, 8)
	if err == nil {
		notify.ReadCurrent(nil, api.ReadCurrent{
			Block:  block,
			Values: values,
		})
		return values, nil
	}
	return nil, merry.Appendf(err, "блок измерения %d", block)
}

func (x worker) switchGasOff() error {
	return x.switchGas(0)
}

func (x worker) switchGas(n int) error {
	var s string
	if n == 0 {
		s = "отключить газ"
	} else {
		s = fmt.Sprintf("подать ПГС%d", n)
	}
	return x.perform(s, func(x worker) error {
		x.ctx = context.Background()
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
		if _, err := req.GetResponse(x.log, x.ReaderGas(), nil); err != nil {
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
		if _, err := req.GetResponse(x.log, x.ReaderGas(), nil); err != nil {
			return err
		}
		x.lastGas = &n
		return nil
	})
}

func blowGas(x worker, n int) error {
	if err := x.performf("включение клапана %d", n)(func(x worker) error {
		return performWithWarn(x, func() error {
			return x.switchGas(n)
		})
	}); err != nil {
		return err
	}
	duration := time.Minute * time.Duration(cfg.Cfg.Gui().BlowGasMinutes)
	return delayf(x, duration, "продувка ПГС%d", n)
}
