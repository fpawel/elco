package daemon

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/pkg/serial-comm/comport"
	"github.com/fpawel/elco/pkg/serial-comm/modbus"
	"github.com/hako/durafmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sort"
	"time"
)

func (x *D) switchGas(n int) error {

	logrus.Infof("переключение газового блока: %d", n)

	err := x.doSwitchGas(n)
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
	notify.Warning(x.w, s)
	if merry.Is(x.hardware.ctx.Err(), context.Canceled) {
		return err
	}
	logrus.Warnf("проигнорирована ошибка связи с газовым блоком при переключении %d: %v", n, err)

	return nil
}

func (x *D) doSwitchGas(n int) error {

	req := modbus.Req{
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

	responseReader := comport.Comm{
		Port:   x.portGas,
		Config: x.cfg.Predefined().ComportGas,
		Ctx:    x.hardware.ctx,
	}

	if !x.portGas.Opened() {
		if err := x.portGas.Open(x.cfg.User().ComportGas, 9600, 0, context.Background()); err != nil {
			return err
		}
	}

	if _, err := responseReader.GetResponse(req.Bytes(), nil); err != nil {
		return err
	}

	req = modbus.Req{
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

	logrus.Warnf("установить расход % X", req.Bytes())

	if _, err := responseReader.GetResponse(req.Bytes(), nil); err != nil {
		return err
	}

	return nil
}

func (x *D) blowGas(nGas int) error {
	if err := x.switchGas(nGas); err != nil {
		return err
	}
	timeMinutes := x.cfg.Predefined().BlowGasMinutes
	return x.delay(fmt.Sprintf("продувка ПГС%d", nGas), time.Minute*time.Duration(timeMinutes))
}

func (x *D) delay(what string, duration time.Duration) error {
	t := time.Now()
	logrus.Infof("%s: %s, начало", what, durafmt.Parse(duration))
	err := x.doDelay(what, duration)
	if err == nil {
		logrus.Infof("%s: %s: выполнено без ошибок", what, durafmt.Parse(duration))
	}
	return merry.Appendf(err, "%s: %s: %s", what, durafmt.Parse(duration), durafmt.Parse(time.Since(t)))

}

func (x *D) doDelay(what string, duration time.Duration) error {

	ctx, skipDelay := context.WithTimeout(x.hardware.ctx, duration)

	t := time.Now()
	x.hardware.skipDelay = func() {
		skipDelay()
		logrus.Warnf("%s %s: задержка прервана: %s", what, durafmt.Parse(duration), durafmt.Parse(time.Since(t)))
	}

	notify.Delay(x.w, api.DelayInfo{
		Run:         true,
		What:        what,
		TimeSeconds: int(duration.Seconds()),
	})

	defer func() {
		notify.Delay(x.w, api.DelayInfo{Run: false})
	}()
	for {
		productsWithSerials, err := data.GetLastPartyProducts(x.dbProducts,
			data.ProductsFilter{
				WithSerials:    true,
				WithProduction: true})
		if err != nil {
			return err
		}
		if len(productsWithSerials) == 0 {
			return merry.New("фоновый опрос: не выбрано ни одного прибора")
		}
		for _, products := range GroupProductsByBlocks(productsWithSerials) {

			if ctx.Err() != nil {
				return nil
			}

			if x.hardware.ctx.Err() != nil {
				return x.hardware.ctx.Err()
			}

			block := products[0].Place / 8

			_, err := x.readBlockMeasure(block, ctx)

			if err == nil {
				continue
			}

			if ctx.Err() != nil {
				return nil
			}

			if x.hardware.ctx.Err() != nil {
				return x.hardware.ctx.Err()
			}

			notify.Warningf(x.w, "фоновый опрос: блок измерения %d: %v", block+1, err)

			if merry.Is(x.hardware.ctx.Err(), context.Canceled) {
				return err
			}

			logrus.Warnf("%s: фоновый опрос: проигнорирована ошибка связи с блоком измерительным %d: %v", what, block, err)

			continue
		}
	}
}

func (x *D) setupTemperature(temperature data.Temperature) error {
	notify.Warningf(x.w, `Установите в термокамере температуру %v⁰C. Нажмите \"Ok\" чтобы перейти к выдержке на температуре %v⁰C.`,
		temperature, temperature)
	duration := time.Minute * time.Duration(x.cfg.Predefined().HoldTemperatureMinutes)
	return x.delay(fmt.Sprintf("выдержка термокамеры: %v⁰C", temperature), duration)
}

func (x *D) determineTemperature(temperature data.Temperature) error {

	if err := x.setupTemperature(temperature); err != nil {
		return err
	}

	if err := x.blowGas(1); err != nil {
		return err
	}

	if err := x.determineProductsTemperatureCurrents(temperature, data.Fon); err != nil {
		return err
	}

	if err := x.blowGas(3); err != nil {
		return err
	}

	if err := x.determineProductsTemperatureCurrents(temperature, data.Sens); err != nil {
		return err
	}

	if err := x.blowGas(1); err != nil {
		return err
	}

	if temperature == 20 {
		if err := x.readAndSaveProductsCurrents("i13"); err != nil {
			return merry.WithMessagef(err, "снятие возврата НКУ")
		}
	}

	if err := x.switchGas(0); err != nil {
		return err
	}

	return nil
}

func (x *D) determineMainError() error {

	for i, pt := range data.MainErrorPoints {
		what := fmt.Sprintf("%d: ПГС%d: снятие основной погрешности", i+1, pt.Code())

		notify.Status(x.w, what)

		if err := x.blowGas(pt.Code()); err != nil {
			return err
		}

		if err := x.readAndSaveProductsCurrents(pt.Field()); err != nil {
			return errors.Wrap(err, what)
		}
	}
	return nil
}

func (x *D) determineProductsTemperatureCurrents(temperature data.Temperature, scale data.ScaleType) error {
	return x.readAndSaveProductsCurrents(data.TemperatureScaleField(temperature, scale))
}

func (x *D) readAndSaveProductsCurrents(field string) error {
	logrus.Infof("снятие %q: начало", field)
	err := x.doReadAndSaveProductsCurrents(field)
	if err == nil {
		logrus.Infof("снятие %q: успешно", field)
		return nil
	}
	return merry.WithValue(err, "field", field).Append("снятие")
}

func (x *D) doReadAndSaveProductsCurrents(field string) error {

	productsToWork, err := data.GetLastPartyProducts(x.dbProducts, data.ProductsFilter{
		WithSerials:    true,
		WithProduction: true,
	})
	if err != nil {
		return err
	}
	if len(productsToWork) == 0 {
		return merry.New("не выбрано ни одного прибора в данной партии")
	}
	blockProducts := GroupProductsByBlocks(productsToWork)
	for _, products := range blockProducts {
		block := products[0].Place / 8

		values, err := x.readBlockMeasure(block, x.hardware.ctx)
		for ; err != nil; values, err = x.readBlockMeasure(block, x.hardware.ctx) {
			if merry.Is(err, context.Canceled) {
				return err
			}
			notify.Warning(x.w, fmt.Sprintf("блок измерения %d: %v", block+1, err))
			if x.hardware.ctx.Err() == context.Canceled {
				return err
			}
		}

		for _, p := range products {
			n := p.Place % 8

			logrus.WithFields(logrus.Fields{
				"product_id": p.ProductID,
				"place":      data.FormatPlace(p.Place),
				"value":      values[n],
				"field":      field,
			}).Info("сохранение")

			if err := data.SetProductValue(x.dbxProducts, p.ProductID, field, values[n]); err != nil {
				return merry.WithValue(err, "product_id", p.ProductID).
					WithValue("place", data.FormatPlace(p.Place)).
					WithValue("value", values[n]).
					WithValue("field", field).Append("сохранение")
			}
		}
	}

	var party data.Party
	if err := data.GetLastParty(x.dbProducts, &party); err != nil {
		return err
	}
	if err := data.GetPartyProducts(x.dbProducts, &party); err != nil {
		return err
	}
	notify.LastPartyChanged(x.w, party)

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

func (x *D) readBlockMeasure(block int, ctx context.Context) ([]float64, error) {

	values, err := modbus.Read3BCDValues(x.commMeasurer(ctx), modbus.Addr(block+101), 0, 8)

	switch err {

	case nil:
		notify.ReadCurrent(x.w, api.ReadCurrent{
			Block:  block,
			Values: values,
		})
		return values, nil

	default:
		return nil, merry.WithValue(err, "block", block)
	}
}

func (x *D) commMeasurer(ctx context.Context) comport.Comm {
	return comport.Comm{
		Port:   x.portMeasurer,
		Config: x.cfg.Predefined().ComportMeasurer,
		Ctx:    ctx,
	}
}

func init() {
	merry.RegisterDetail("Запрос", "request")
	merry.RegisterDetail("Ответ", "response")
	merry.RegisterDetail("Длительность ожидания", "duration")
	merry.RegisterDetail("Порт", "port")
	merry.RegisterDetail("Прибор", "device")
	merry.RegisterDetail("Блок измерительный", "block")
	merry.RegisterDetail("Длительность ожидания статуса", "status_timeout")
	merry.RegisterDetail("Место", "place")
	merry.RegisterDetail("Код статуса", "status")
	merry.RegisterDetail("Адрес", "addr")

}
