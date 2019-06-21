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
	"github.com/hako/durafmt"
	"github.com/pkg/errors"
	"github.com/powerman/structlog"
	"github.com/sirupsen/logrus"
	"math"
	"sort"
	"time"
)

func switchGas(n int) error {

	logrus.Infof("переключение газового блока: %d", n)

	err := doSwitchGas(n)
	if err == nil {
		logrus.Infof("газового блока переключен: %d", n)
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
	notify.Warning(log, s)
	if merry.Is(hardware.ctx.Err(), context.Canceled) {
		return err
	}
	logrus.Warnf("проигнорирована ошибка связи с газовым блоком при переключении %d: %v", n, err)

	return nil
}

func doSwitchGas(n int) error {

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

	if !portGas.Opened() {
		if err := portGas.Open(cfg.Cfg.User().ComportGas); err != nil {
			return err
		}
	}

	log := gohelp.LogWithKeys(log, "пневмоблок", n)

	if _, err := req.GetResponse(log, gasBlockReader(), nil); err != nil {
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

	log = gohelp.LogWithKeys(log, "пневмоблок", "расход")

	if _, err := req.GetResponse(log, gasBlockReader(), nil); err != nil {
		return err
	}

	return nil
}

func blowGas(nGas int) error {
	if err := switchGas(nGas); err != nil {
		return err
	}
	timeMinutes := cfg.Cfg.Predefined().BlowGasMinutes
	return delay(fmt.Sprintf("продувка ПГС%d", nGas), time.Minute*time.Duration(timeMinutes))
}

func delay(what string, duration time.Duration) error {
	t := time.Now()
	logrus.Infof("%s: %s, начало", what, durafmt.Parse(duration))
	err := doDelay(what, duration)
	if err == nil {
		logrus.Infof("%s: %s: выполнено без ошибок", what, durafmt.Parse(duration))
	}
	return merry.Appendf(err, "%s: %s: %s", what, durafmt.Parse(duration), durafmt.Parse(time.Since(t)))

}

func doDelay(what string, duration time.Duration) error {

	ctx, skipDelay := context.WithTimeout(hardware.ctx, duration)

	t := time.Now()
	hardware.skipDelayFunc = func() {
		skipDelay()
		logrus.Warnf("%s %s: задержка прервана: %s", what, durafmt.Parse(duration), durafmt.Parse(time.Since(t)))
	}

	notify.Delay(log, api.DelayInfo{
		Run:         true,
		What:        what,
		TimeSeconds: int(duration.Seconds()),
	})

	defer func() {
		notify.Delay(log, api.DelayInfo{Run: false})
	}()
	for {
		products := data.GetLastPartyProducts(data.WithSerials, data.WithProduction)

		if len(products) == 0 {
			return merry.New("фоновый опрос: не выбрано ни одного прибора")
		}
		for _, products := range GroupProductsByBlocks(products) {

			if ctx.Err() != nil {
				return nil
			}

			if hardware.ctx.Err() != nil {
				return hardware.ctx.Err()
			}

			block := products[0].Place / 8

			_, err := readBlockMeasure(
				gohelp.LogWithKeys(log, "фоновый_опрос", fmt.Sprintf("%s %s", what, durafmt.Parse(duration))),
				block, ctx)

			if err == nil {
				continue
			}

			if ctx.Err() != nil {
				return nil
			}

			if hardware.ctx.Err() != nil {
				return hardware.ctx.Err()
			}

			notify.Warningf(log, "фоновый опрос: блок измерения %d: %v", block+1, err)

			if merry.Is(hardware.ctx.Err(), context.Canceled) {
				return err
			}

			logrus.Warnf("%s: фоновый опрос: проигнорирована ошибка связи с блоком измерительным %d: %v", what, block, err)

			continue
		}
	}
}

func doSetupTemperature(destinationTemperature float64) error {
	// запись уставки
	if err := ktx500.WriteDestination(destinationTemperature); err != nil {
		return err
	}
	// включение термокамеры
	if err := ktx500.WriteOnOff(true); err != nil {
		return err
	}

	// установка компрессора
	if err := ktx500.WriteCoolOnOff(destinationTemperature < 50); err != nil {
		return err
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
				gohelp.LogWithKeys(log, "фоновый_опрос",
					fmt.Sprintf("установка температуры %v⁰C", destinationTemperature)),
				block, hardware.ctx); err != nil {
				return err
			}
		}
	}
}

func setupTemperature(temperature data.Temperature) error {

	err := doSetupTemperature(float64(temperature))
	if err != nil {
		if !merry.Is(err, ktx500.Err) {
			return err
		}
		notify.Warningf(log, `Не удалось установить температуру: %v⁰C: %v`, temperature, err)
		if merry.Is(hardware.ctx.Err(), context.Canceled) {
			return err
		}

		logrus.Warnf("установка тепературы %v⁰C, фоновый опрос: проигнорирована ошибка связи с термокамерой: %v",
			temperature, err)
	}

	duration := time.Minute * time.Duration(cfg.Cfg.Predefined().HoldTemperatureMinutes)
	return delay(fmt.Sprintf("выдержка термокамеры: %v⁰C", temperature), duration)
}

func determineTemperature(temperature data.Temperature) error {

	if err := setupTemperature(temperature); err != nil {
		return err
	}

	defer func() {

		// выключение термокамеры
		log.ErrIfFail(func() error {
			return ktx500.WriteOnOff(false)
		})
		// выключение компрессора
		log.ErrIfFail(func() error {
			return ktx500.WriteCoolOnOff(false)
		})
		// выключение газового блока
		log.ErrIfFail(func() error {
			return switchGas(0)
		})

	}()

	if err := blowGas(1); err != nil {
		return err
	}

	if err := determineProductsTemperatureCurrents(temperature, data.Fon); err != nil {
		return err
	}

	if err := blowGas(3); err != nil {
		return err
	}

	if err := determineProductsTemperatureCurrents(temperature, data.Sens); err != nil {
		return err
	}

	if err := blowGas(1); err != nil {
		return err
	}

	if temperature == 20 {
		if err := readAndSaveProductsCurrents("i13"); err != nil {
			return merry.WithMessagef(err, "снятие возврата НКУ")
		}
	}

	return nil
}

func determineMainError() error {

	for i, pt := range data.MainErrorPoints {
		what := fmt.Sprintf("%d: ПГС%d: снятие основной погрешности", i+1, pt.Code())

		notify.Status(log, what)

		if err := blowGas(pt.Code()); err != nil {
			return err
		}

		if err := readAndSaveProductsCurrents(pt.Field()); err != nil {
			return errors.Wrap(err, what)
		}
	}
	return nil
}

func determineProductsTemperatureCurrents(temperature data.Temperature, scale data.ScaleType) error {
	return readAndSaveProductsCurrents(data.TemperatureScaleField(temperature, scale))
}

func readAndSaveProductsCurrents(field string) error {
	logrus.Infof("снятие %q: начало", field)
	err := doReadAndSaveProductsCurrents(field)
	if err == nil {
		logrus.Infof("снятие %q: успешно", field)
		return nil
	}
	return merry.WithValue(err, "field", field).Append("снятие")
}

func doReadAndSaveProductsCurrents(field string) error {

	productsToWork := data.GetLastPartyProducts(data.WithSerials, data.WithProduction)

	if len(productsToWork) == 0 {
		return merry.New("не выбрано ни одного прибора в данной партии")
	}
	logrus.Infof("снятие %q: %s", field, formatProducts(productsToWork))

	log := gohelp.LogWithKeys(log, "снятие", field)

	blockProducts := GroupProductsByBlocks(productsToWork)
	for _, products := range blockProducts {
		block := products[0].Place / 8

		values, err := readBlockMeasure(log, block, hardware.ctx)
		for ; err != nil; values, err = readBlockMeasure(log, block, hardware.ctx) {
			if merry.Is(err, context.Canceled) {
				return err
			}
			notify.Warning(log, fmt.Sprintf("блок измерения %d: %v", block+1, err))
			if hardware.ctx.Err() == context.Canceled {
				return err
			}
		}

		for _, p := range products {
			n := p.Place % 8
			args := []interface{}{
				"product_id", p.ProductID,
				"place", data.FormatPlace(p.Place),
				"field", field,
				"value", values[n],
			}
			if err := data.SetProductValue(p.ProductID, field, values[n]); err != nil {
				return log.Err(merry.Append(err, "не удалось сохранить"), args...)
			}
			log.Info("сохраненено", args...)
		}
	}
	party := data.GetLastParty(data.WithProducts)
	notify.LastPartyChanged(log, party)
	return nil

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

func readBlockMeasure(logger *structlog.Logger, block int, ctx context.Context) ([]float64, error) {

	log := gohelp.LogWithKeys(logger, "блок", block)

	values, err := modbus.Read3BCDs(log, measurerReader(ctx), modbus.Addr(block+101), 0, 8)

	switch err {

	case nil:
		notify.ReadCurrent(log, api.ReadCurrent{
			Block:  block,
			Values: values,
		})
		return values, nil

	default:
		return nil, merry.WithValue(err, "block", block)
	}
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
