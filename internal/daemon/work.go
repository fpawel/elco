package daemon

import (
	"context"
	"fmt"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/goutils/serial-comm/comport"
	"github.com/fpawel/goutils/serial-comm/modbus"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"log"
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
		if err == context.DeadlineExceeded {
			err = errors.New("нет ответа от газового блока: " + x.port.gas.Dump())
		}
		return errors.Wrapf(err, "газовый блок: %s", x.port.gas.Dump())
	}
	return nil
}

func (x *D) doSwitchGas(n int) error {
	c := x.sets.Config()
	if !x.port.gas.Opened() {
		if err := x.port.gas.Open(c.Comport.GasSwitcher, 9600, 0, context.Background()); err != nil {
			return err
		}
	}
	req := modbus.Request{
		Addr:                5,
		ProtocolCommandCode: 0x10,
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

	req = modbus.Request{
		Addr:                1,
		ProtocolCommandCode: 6,
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

func (x *D) readMeasure(place int) ([]float64, error) {

	c := x.sets.Config()

	switch values, err := modbus.Read3BCDValues(comport.Comm{
		Port:   x.port.measurer,
		Config: c.Measurer,
	}, modbus.Addr(place+101), 0, 8); err {

	case nil:

		notify.ReadCurrent(x.w, api.ReadCurrent{
			Place:  place,
			Values: values,
		})
		return values, nil

	case context.Canceled:
		return nil, context.Canceled

	case context.DeadlineExceeded:

		return nil, errors.Errorf("блок измерения №%d: не отечает: %s",
			place+1,
			x.port.measurer.Dump())

	default:
		return nil, errors.Errorf("блок измерения №%d: %+v: %s",
			place+1, err,
			x.port.measurer.Dump())
	}
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
		return errors.Wrap(err, "снятие возврата НКУ")
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
	products, blocks := x.c.LastParty().ProductsWithSerials()

	for i := range blocks {
		if blocks[i] {
			values, err := x.readMeasure(i)
			if err != nil {
				return err
			}
			for n := 0; n < 8; n++ {
				p := products[i*8+n]
				fields["product_id"] = p.ProductID
				fields["place"] = p.Place
				fields["value"] = values[n]
				logrus.WithFields(fields).Info("save product current")
				f(p, values[n])
				x.c.SaveProduct(p)
			}
		}
	}
	notify.LastPartyChanged(x.w, x.c.LastParty().Party())
	return nil
}