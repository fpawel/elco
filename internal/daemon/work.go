package daemon

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/goutils/serial-comm/modbus"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"log"
	"sort"
	"time"
)

func (x *D) switchGas(n int) error {
	err := x.doSwitchGas(n)
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
	notify.Warning(x.w, s)
	if x.hardware.ctx.Err() == context.Canceled {
		return err
	}
	return nil
}

func (x *D) getResponseGasSwitcher(b []byte) error {
	c := x.sets.Config()
	if _, err := x.port.gas.GetResponse(b, c.GasSwitcher); err != nil {
		w := x.port.gas.LastWork()
		if err == context.DeadlineExceeded {
			err = merry.Errorf("нет ответа от газового блока: %s", w.FormatRequest())
		}
		return merry.Appendf(err, "газовый блок: %s", w.FormatResponse())
	}
	return nil
}

func (x *D) doSwitchGas(n int) error {
	c := x.sets.Config()
	if !x.port.gas.Opened() {
		x.port.gas.SetLog(true)
		if err := x.port.gas.Open(c.Comport.GasSwitcher, 9600, 0, context.Background()); err != nil {
			return err
		}
	}
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
		log.Panicf("wrong gas code: %d", n)
	}

	if err := x.getResponseGasSwitcher(req.Bytes()); err != nil {
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

	if err := x.getResponseGasSwitcher(req.Bytes()); err != nil {
		return err
	}

	logrus.WithField("code", n).Debug("gas block")

	return nil
}

func (x *D) blowGas(nGas int) error {
	if err := x.switchGas(nGas); err != nil {
		return err
	}
	timeMinutes := x.sets.Config().Work.BlowGasMinutes
	x.delay(fmt.Sprintf("Продувка ПГС%d", nGas), time.Minute*time.Duration(timeMinutes))
	return x.hardware.ctx.Err()
}

func (x *D) delay(what string, duration time.Duration) {

	var ctx context.Context
	ctx, x.hardware.skipDelay = context.WithCancel(x.hardware.ctx)

	t := time.After(duration)

	notify.Delay(x.w, api.DelayInfo{
		Run:         true,
		What:        what,
		TimeSeconds: int(duration.Seconds()),
	})

	defer notify.Delay(x.w, api.DelayInfo{Run: false})

	for {
		select {

		case <-ctx.Done():
			return

		case <-t:
			return

		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (x *D) setupTemperature(temperature data.Temperature) error {
	notify.Warningf(x.w, `Установите в термокамере температуру %v⁰C. 
Нажмите \"Ok\" чтобы перейти к выдержке на температуре %v⁰C.`, temperature, temperature)
	duration := time.Minute * time.Duration(x.sets.Config().Work.HoldTemperatureMinutes)
	x.delay(fmt.Sprintf("Выдержка термокамеры: %v⁰C", temperature), duration)
	return x.hardware.ctx.Err()
}

func (x *D) determineNKU2() error {
	if err := x.setupTemperature(20); err != nil {
		return err
	}
	if err := x.blowGas(1); err != nil {
		return err
	}
	m := logrus.Fields{
		"return_NKU": struct{}{},
	}
	if err := x.determineProductsCurrents(m, func(p *data.Product, value float64) {
		p.I13.Valid = true
		p.I13.Float64 = value
	}); err != nil {
		return merry.WithMessagef(err, "снятие возврата НКУ")
	}
	return nil
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
		m := logrus.Fields{
			"main_error": pt,
		}
		if err := x.determineProductsCurrents(m, func(p *data.Product, value float64) {
			p.SetMainErrorCurrent(pt, value)
		}); err != nil {
			return errors.Wrap(err, what)
		}
	}
	return nil
}

func (x *D) determineProductsTemperatureCurrents(temperature data.Temperature, scale data.ScaleType) error {
	return x.determineProductsCurrents(logrus.Fields{
		"scale":       scale,
		"temperature": temperature,
	}, func(p *data.Product, value float64) {
		p.SetCurrent(temperature, scale, value)
	})
}

func (x *D) determineProductsCurrents(fields logrus.Fields, f func(*data.Product, float64)) error {

	for _, products := range GroupProductsByBlocks(x.c.LastParty().ProductsWithSerials()) {
		block := products[0].Place / 8
		values, err := x.readBlockMeasure(block)
		if err != nil {
			for k, v := range fields {
				err = merry.WithValue(err, k, v)
			}
			return err
		}
		for _, p := range products {
			n := p.Place % 8
			fields["product_id"] = p.ProductID
			fields["place"] = p.Place
			fields["value"] = values[n]
			logrus.WithFields(fields).Info("save product current")
			f(p, values[n])
			x.c.SaveProduct(p)
		}
	}
	notify.LastPartyChanged(x.w, x.c.LastParty().Party())
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
