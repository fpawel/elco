package app

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/elco/internal/cfg"
	"os"
	"time"
)

func switchGasWithWarn(x worker, n int) error {
	var s string
	if n == 0 {
		s = "отключить газ"
	} else {
		s = fmt.Sprintf("подать ПГС%d", n)
	}
	return x.perform(s, func(x worker) error {
		return x.performWithWarn(func() error {
			return switchGasWithoutWarn(x, n)
		})
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
	return x.performf("продувка ПГС%d", n)(func(x worker) error {
		if err := switchGasWithWarn(x, n); err != nil {
			return err
		}
		duration := time.Minute * time.Duration(cfg.Cfg.User().BlowGasMinutes)
		return delayf(x, duration, "продувка ПГС%d", n)
	})
}
