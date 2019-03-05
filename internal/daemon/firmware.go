package daemon

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/pkg/intrng"
	"github.com/fpawel/elco/pkg/serial-comm/comm"
	"github.com/fpawel/elco/pkg/serial-comm/modbus"
	"github.com/hako/durafmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/reform.v1"
	"sort"
	"time"
)

func (x *D) writePartyFirmware() error {

	startTime := time.Now()
	var party data.Party
	if err := data.GetLastParty(x.dbProducts, &party); err != nil {
		return err
	}

	products, err := data.GetLastPartyProducts(x.dbProducts, data.ProductsFilter{
		WithProduction: true,
		WithSerials:    true,
	})
	if err != nil {
		return err
	}

	if len(products) == 0 {
		return merry.New("не выбрано ни одного прибора")
	}

	logrus.Info(formatProducts(products))

	placeBytes := map[int][]byte{}

	blockProducts := GroupProductsByBlocks(products)
	for _, products := range blockProducts {
		if err := x.writeBlock(products, placeBytes); err != nil {
			return err
		}
	}
	logrus.Infof("запись партии завершена: %s, %s",
		formatProducts(products),
		durafmt.Parse(time.Since(startTime)))

	startTime = time.Now()
	for _, products := range blockProducts {
		var places []int
		for _, p := range products {
			places = append(places, p.Place)
		}
		sort.Ints(places)
		if err := x.verifyProductsFirmware(places, placeBytes); err != nil {
			return merry.Appendf(err, "считывание: %s",
				durafmt.Parse(time.Since(startTime)))
		}
	}
	logrus.Infof("считывание партии завершено: %s, %s",
		formatProducts(products), durafmt.Parse(time.Since(startTime)))

	if err = data.GetPartyProducts(x.dbProducts, &party); err != nil {
		return err
	}
	notify.LastPartyChanged(x.w, party)
	return nil
}

func (x *D) verifyProductsFirmware(places []int, placeBytes map[int][]byte) error {
	for _, place := range places {
		b, err := x.readPlaceFirmware(place)
		if err != nil {
			return err
		}
		for _, c := range firmwareAddresses {
			read := b[c.addr1 : c.addr2+1]
			calc := placeBytes[place][c.addr1 : c.addr2+1]
			if !compareBytes(read, calc) {
				return merry.Errorf(
					"место %s: не совпадают данные по адресам %X...%X",
					data.FormatPlace(place), c.addr1, c.addr2).
					WithValue("расчитано", fmt.Sprintf("% X", read)).
					WithValue("записано", fmt.Sprintf("% X", calc))
			}
		}
	}
	return nil
}

func (x *D) writeBlock(products []*data.Product, placeBytes map[int][]byte) error {
	startTime := time.Now()

	block := products[0].Place / 8
	var placesMask byte
	for _, p := range products {
		place := byte(p.Place) % 8
		placesMask |= 1 << place
	}

	var placesInBlock []int
	for _, p := range products {
		placesInBlock = append(placesInBlock, p.Place%8)
	}
	sort.Ints(placesInBlock)

	for _, p := range products {
		prodInfo := new(data.ProductInfo)
		if err := x.dbProducts.FindByPrimaryKeyTo(prodInfo, p.ProductID); err != nil {
			return err
		}
		firmware, err := prodInfo.Firmware()
		if err != nil {
			return merry.Appendf(err, "расчёт прошивки ЭХЯ не удался: %s", prodInfo.String2())
		}
		placeBytes[p.Place] = firmware.Bytes()
	}

	doAddresses := func(p *data.Product, b data.FirmwareBytes, addr1, addr2 uint16) error {

		placeInBlock := p.Place % 8

		x.logFields["адрес_начала_куска"] = addr1
		x.logFields["адрес_конца_куска"] = addr2
		x.logFields["количество_байт_куска"] = addr2 + 1 - addr1
		x.logFields["количество_байт_куска"] = addr2 + 1 - addr1
		defer delete(x.logFields, "адрес_начала_куска")
		defer delete(x.logFields, "адрес_конца_куска")
		defer delete(x.logFields, "количество_байт_куска")

		//logrus.Infof("место %s: прошивка куска % X...% X", data.FormatPlace(p.Place), addr1, addr2)

		if err := x.sendDataToWrite42(block, placeInBlock, b[addr1:addr2+1]); err != nil {
			return err
		}
		return nil
	}

	for i, c := range firmwareAddresses {
		for _, p := range products {
			if err := doAddresses(p, placeBytes[p.Place], c.addr1, c.addr2); err != nil {
				return err
			}
		}
		if err := x.writePreparedDataToFlash(block, placesMask, c.addr1, int(c.addr2-c.addr1+1)); err != nil {
			return err
		}
		if err := x.waitStatus45(block, placesMask); err != nil {
			return err
		}

		if i < len(firmwareAddresses)-1 {
			time.Sleep(time.Duration(x.cfg.Predefined().ReadRangeDelayMillis) * time.Millisecond)
		}
	}

	for _, p := range products {
		p.Firmware = placeBytes[p.Place]
		if err := x.dbProducts.Save(p); err != nil {
			return err
		}
	}

	logrus.WithField("places_mask", fmt.Sprintf("%08b", placesMask)).
		Infof("запись блока %d %s завершена: %s",
			block, intrng.Format(placesInBlock), durafmt.Parse(time.Since(startTime)))
	return nil
}

func (x *D) readPlaceFirmware(place int) ([]byte, error) {
	startTime := time.Now()
	block := place / 8
	placeInBlock := place % 8
	b := make([]byte, data.FirmwareSize)
	for i := range b {
		b[i] = 0xff
	}
	for i, c := range firmwareAddresses {
		count := c.addr2 - c.addr1 + 1
		req := modbus.Req{
			Addr:     modbus.Addr(block) + 101,
			ProtoCmd: 0x44,
			Data: []byte{
				byte(placeInBlock + 1),
				byte(x.cfg.User().ChipType),
				byte(c.addr1 >> 8),
				byte(c.addr1),
				byte(count >> 8),
				byte(count),
			},
		}
		resp, err := x.measurerReader(x.hardware.ctx).GetResponse(req.Bytes(), func(request, response []byte) error {
			if len(response) != 10+int(count) {
				return comm.ErrProtocol.Here().WithMessagef("ожидалось %d байт ответа, получено %d",
					10+int(count), len(response))
			}
			return nil
		})
		if err != nil {
			return nil, err
		}

		copy(b[c.addr1:c.addr1+count], resp[8:8+count])
		if i < len(firmwareAddresses)-1 {
			time.Sleep(time.Duration(
				x.cfg.Predefined().ReadRangeDelayMillis) *
				time.Millisecond)
		}
	}
	logrus.Infof("считана прошивка метса %s: %d байт: %s",
		data.FormatPlace(place), len(b), durafmt.Parse(time.Since(startTime)))
	return b, nil
}

func (x *D) writePlaceFirmware(place int, bytes []byte) error {
	block := place / 8
	placeInBlock := place % 8
	placesMask := byte(1) << byte(placeInBlock)
	startTime := time.Now()
	doAddresses := func(addr1, addr2 uint16) error {
		x.logFields["адрес_начала_куска"] = addr1
		x.logFields["адрес_конца_куска"] = addr2
		x.logFields["количество_байт_куска"] = addr2 + 1 - addr1
		defer delete(x.logFields, "адрес_начала_куска")
		defer delete(x.logFields, "адрес_конца_куска")
		defer delete(x.logFields, "количество_байт_куска")

		//logrus.WithFields(logrus.Fields{}).Infof("запись куска прошивки ЭХЯ: %d...%d, %d байт", addr1, addr2, addr2+1-addr1)

		if err := x.sendDataToWrite42(block, placeInBlock, bytes[addr1:addr2+1]); err != nil {
			return err
		}

		if err := x.writePreparedDataToFlash(block, placesMask, addr1, int(addr2-addr1+1)); err != nil {
			return err
		}
		time.Sleep(time.Second)

		if err := x.waitStatus45(block, placesMask); err != nil {
			return err
		}
		return nil
	}

	for _, c := range firmwareAddresses {
		if err := doAddresses(c.addr1, c.addr2); err != nil {
			return err
		}
	}
	//logrus.Info("запись прошивки ЭХЯ выполнена успешно")

	var p data.Product
	switch err := data.GetLastPartyProductAtPlace(x.dbProducts, place, &p); err {
	case nil:
		p.Firmware = bytes
		if err := x.dbProducts.Save(&p); err != nil {
			return err
		}
	case reform.ErrNoRows, sql.ErrNoRows:
		return nil
	default:
		return err
	}
	logrus.Infof("запись места: %s, %s", data.FormatPlace(place),
		durafmt.Parse(time.Since(startTime)))
	return nil
}

func (x *D) waitStatus45(block int, placesMask byte) error {
	t := time.Duration(x.cfg.Predefined().StatusTimeoutSeconds) * time.Second
	ctx, _ := context.WithTimeout(x.hardware.ctx, t)
	for {

		select {
		case <-ctx.Done():
			response, err := x.readStatus45(block)
			if err != nil {
				return err
			}
			status := response[2:10]

			for i, b := range status {
				if (1<<byte(i))&placesMask != 0 && b != 0 {
					return merry.Errorf(
						"%s: таймаут %s, статус[%d]=%X",
						data.FormatPlace(block*8+i), durafmt.Parse(t), i, b)
				}
			}
			return nil

		default:
			response, err := x.readStatus45(block)
			if err != nil {
				return err
			}
			status := response[2:10]
			statusOk := true
			for i, b := range status {
				if (1<<byte(i))&placesMask != 0 {
					if b == 0 {
						continue
					}
					statusOk = false
					if b != 0xB2 {
						return merry.Errorf("место %s: статус[%d]=%X",
							data.FormatPlace(block*8+i), i, b)
					}
				}
			}
			if statusOk {
				return nil
			}
		}
	}
}

func (x *D) readStatus45(block int) ([]byte, error) {
	return x.measurerReader(x.hardware.ctx).GetResponse(modbus.Req{
		Addr:     modbus.Addr(block) + 101,
		ProtoCmd: 0x45,
	}.Bytes(), func(request, response []byte) error {
		if len(response) != 12 {
			return comm.ErrProtocol.Here().WithMessagef("ожидалось 12 байт ответа, получено %d", len(response))
		}
		return nil
	})
}

func (x *D) writePreparedDataToFlash(block int, placesMask byte, addr uint16, count int) error {
	req := modbus.Req{
		Addr:     modbus.Addr(block) + 101,
		ProtoCmd: 0x43,
		Data: []byte{
			placesMask,
			byte(x.cfg.User().ChipType),
			byte(addr >> 8),
			byte(addr),
			byte(count >> 8),
			byte(count),
		},
	}
	request := req.Bytes()

	_, err := x.measurerReader(x.hardware.ctx).GetResponse(request, func(request, response []byte) error {
		if !compareBytes(response, req.Bytes()) {
			return merry.New("запрос не равен ответу")
		}
		return nil
	})
	return err
}

func (x *D) sendDataToWrite42(block, placeInBlock int, b []byte) error {
	req := modbus.Req{
		Addr:     modbus.Addr(block) + 101,
		ProtoCmd: 0x42,
		Data: append([]byte{
			byte(placeInBlock + 1),
			byte(len(b) >> 8),
			byte(len(b)),
		}, b...),
	}
	request := req.Bytes()
	_, err := x.measurerReader(x.hardware.ctx).GetResponse(request, func(request, response []byte) error {
		if len(response) != 7 {
			return merry.Errorf("длина ответа %d не равна 7", len(response))
		}
		if !compareBytes(response[:5], req.Bytes()[:5]) {
			return merry.Errorf("% X != % X", response[:5], request[:5])
		}
		return nil
	})

	return err
}

func compareBytes(x, y []byte) bool {
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

var firmwareAddresses = []struct{ addr1, addr2 uint16 }{
	{0, 512},
	{1024, 1535},
	{1536, 1600},
	{1792, 1810},
	{1824, 1831},
}
