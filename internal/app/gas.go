package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/gohelp"
	"github.com/powerman/structlog"
	"os"
	"time"
)

func switchGasWithWarn(log *structlog.Logger, n int) error {

	err := switchGasWithoutWarn(log, n)
	if err == nil {
		return nil
	}

	s := "Не удалось "
	if n == 0 {
		s += "отключить газ"
	} else {
		s += fmt.Sprintf("подать ПГС%d", n)
	}

	s += ": " + err.Error() + ".\n\n"

	if n == 0 {
		s += "Отключите газ"
	} else {
		s += fmt.Sprintf("Подайте ПГС%d", n)
	}
	s += " вручную."
	notify.WarningSync(log, s)
	if merry.Is(ctxWork.Err(), context.Canceled) {
		return err
	}
	log.Warn("проигнорирована ошибка связи с газовым блоком", "gas", n, "error", err)

	return nil
}

func switchGasWithoutWarn(log *structlog.Logger, n int) error {

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

	log = gohelp.LogPrependSuffixKeys(log, "gas", n)

	if os.Getenv("ELCO_DEBUG_NO_HARDWARE") == "true" {
		log.Warn("skip because ELCO_DEBUG_NO_HARDWARE==true")
		return nil
	}

	log.Info("переключение клапана")

	if _, err := req.GetResponse(log, ctxWork, portGas, nil); err != nil {
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

	log.Info("установка расхода")

	if _, err := req.GetResponse(log, ctxWork, portGas, nil); err != nil {
		return err
	}

	return nil
}

func blowGas(log *structlog.Logger, nGas int) error {
	if err := switchGasWithWarn(log, nGas); err != nil {
		return err
	}
	duration := time.Minute * time.Duration(cfg.Cfg.User().BlowGasMinutes)
	return delay(log, fmt.Sprintf("продувка ПГС%d", nGas), duration)
}
