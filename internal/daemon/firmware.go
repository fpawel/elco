package daemon

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/pkg/serial-comm/comport"
	"github.com/fpawel/elco/pkg/serial-comm/modbus"
	"github.com/hako/durafmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/reform.v1"
	"sort"
	"time"
)

func (x *D) writePartyFirmware() error {

	party, err := data.GetLastParty(x.db)
	if err != nil {
		return err
	}

	products, err := data.GetLastPartyProducts(x.db, data.ProductsFilter{
		WithProduction: true,
	})
	if err != nil {
		return err
	}

	var places []int
	for _, p := range products {
		places = append(places, p.Place)
	}
	sort.Ints(places)
	logrus.Infof("Запись прошивки партии: %s, %v", party.String2(), places)
	blockProducts := GroupProductsByBlocks(products)
	for _, products := range blockProducts {
		if err := x.writeProductsFirmware(products); err != nil {
			return err
		}
	}
	return nil
}

func (x *D) writeProductsFirmware(products []*data.Product) error {

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

	x.logFields["block"] = block
	x.logFields["places_in_block"] = fmt.Sprintf("%d", placesInBlock)
	x.logFields["places_mask"] = fmt.Sprintf("%08b", placesMask)
	defer func() {
		delete(x.logFields, "block")
		delete(x.logFields, "products")
		delete(x.logFields, "places_mask")

	}()

	logrus.Infof("запись прошивки ячеек блока %d: %v", block, placesInBlock)

	doAddresses := func(p *data.Product, b data.FirmwareBytes, addr1, addr2 uint16) error {
		x.logFields["адрес_начала_куска"] = addr1
		x.logFields["адрес_конца_куска"] = addr2
		x.logFields["количество_байт_куска"] = addr2 + 1 - addr1
		defer delete(x.logFields, "адрес_начала_куска")
		defer delete(x.logFields, "адрес_конца_куска")
		defer delete(x.logFields, "количество_байт_куска")

		placeInBlock := p.Place % 8

		if err := x.sendDataToWriteFlash(block, placeInBlock, b[addr1:addr2+1]); err != nil {
			return err
		}

		if err := x.writePreparedDataToFlash(block, placesMask, addr1, int(addr2-addr1+1)); err != nil {
			return err
		}
		if err := x.waitFirmwareStatus(block, placesMask); err != nil {
			return err
		}
		return nil
	}

	for _, p := range products {

		prodInfo, err := data.GetProductInfoWithID(x.db, p.ProductID)
		if err != nil {
			return err
		}

		logrus.Infoln("расчёт и запись прошивки ЭХЯ:", prodInfo.String2())

		firmware, err := prodInfo.Firmware()
		if err != nil {
			return merry.Appendf(err, "расчёт прошивки ЭХЯ не удался %v", prodInfo)
		}
		b := firmware.Bytes()
		logrus.Infoln("расчитана прошивка ЭХЯ:", firmware.String2())
		for i, c := range firmwareAddresses {
			if err := doAddresses(p, b, c.addr1, c.addr2); err != nil {
				return err
			}
			if i < len(firmwareAddresses)-1 {
				time.Sleep(x.cfg.Predefined().ReadRangeDelayMillis * time.Millisecond)
			}
		}
	}

	return nil
}

func (x *D) readFirmware(place int) ([]byte, error) {

	x.logFields["place"] = place
	defer delete(x.logFields, "place")
	logrus.Info("считывание прошивки ЭХЯ")

	block := place / 8
	placeInBlock := place % 8

	responseReader := comport.Comm{
		Port:   x.portMeasurer,
		Config: x.cfg.Predefined().ComportMeasurer,
	}

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
		resp, err := responseReader.GetResponse(req.Bytes())
		if err != nil {
			return nil, err
		}
		if err = req.CheckResponse(resp); err != nil {
			return nil, x.portMeasurer.WrapError(err)
		}
		if len(resp) != 10+int(count) {
			return nil, x.portMeasurer.Errorf("ожидалось %d байт ответа, получено %d",
				10+int(count), len(resp))
		}
		copy(b[c.addr1:c.addr1+count], resp[8:8+count])
		if i < len(firmwareAddresses)-1 {
			time.Sleep(x.cfg.Predefined().ReadRangeDelayMillis * time.Millisecond)
		}
	}
	logrus.Infof("считана прошивка ЭХЯ: %d байт, % X", len(b), b)
	return b, nil
}

func (x *D) writeFirmware(place int, bytes []byte) error {
	x.logFields["place"] = place
	defer delete(x.logFields, "place")
	logrus.Infof("запись прошивки ЭХЯ: %d байт, % X", len(bytes), bytes)

	block := place / 8
	placeInBlock := place % 8
	placesMask := byte(1) << byte(place)

	doAddresses := func(addr1, addr2 uint16) error {
		x.logFields["адрес_начала_куска"] = addr1
		x.logFields["адрес_конца_куска"] = addr2
		x.logFields["количество_байт_куска"] = addr2 + 1 - addr1
		defer delete(x.logFields, "адрес_начала_куска")
		defer delete(x.logFields, "адрес_конца_куска")
		defer delete(x.logFields, "количество_байт_куска")

		logrus.WithFields(logrus.Fields{}).Infof("запись куска прошивки ЭХЯ: %d...%d, %d байт", addr1, addr2, addr2+1-addr1)

		if err := x.sendDataToWriteFlash(block, placeInBlock, bytes[addr1:addr2+1]); err != nil {
			return err
		}

		if err := x.writePreparedDataToFlash(block, placesMask, addr1, int(addr2-addr1+1)); err != nil {
			return err
		}

		if err := x.waitFirmwareStatus(block, placesMask); err != nil {
			return err
		}
		return nil
	}

	for _, c := range firmwareAddresses {
		if err := doAddresses(c.addr1, c.addr2); err != nil {
			return err
		}
	}
	logrus.Info("запись прошивки ЭХЯ выполнена успешно")

	switch p, err := data.GetLastPartyProductAtPlace(x.db, place); err {
	case nil:
		p.Firmware = bytes
		if err := x.db.Save(&p); err != nil {
			return err
		}
	case reform.ErrNoRows, sql.ErrNoRows:
		return nil
	default:
		return err
	}
	return nil
}

func (x *D) waitFirmwareStatus(block int, placesMask byte) error {

	t := x.cfg.Predefined().StatusTimeoutSeconds * time.Second
	logrus.Infof("прошивка блока %d: ожидание статуса завершения, таймаут %s", block, durafmt.Parse(t))
	ctx, _ := context.WithTimeout(x.hardware.ctx, t)
	for {

		select {
		case <-ctx.Done():
			status, err := x.readFirmwareStatus(block)
			if err != nil {
				return x.portMeasurer.WrapError(err)
			}
			for i, b := range status {
				if (1<<byte(i))&placesMask != 0 && b != 0 {
					return merry.Errorf("не удалось записать прошивку места %d: таймаут %s, статус[%d]=%X",
						durafmt.Parse(t), block*8+i+1, i, b)
				}
			}
			return nil

		default:
			status, err := x.readFirmwareStatus(block)
			if err != nil {
				return x.portMeasurer.WrapError(err)
			}
			statusOk := true
			for i, b := range status {
				if (1<<byte(i))&placesMask != 0 {
					if b == 0 {
						continue
					}
					statusOk = false
					if b != 0xB2 {
						return merry.Errorf("не удалось записать прошивку места %d: статус[%d]=%X",
							block*8+i+1, i, b)
					}
				}
			}
			if statusOk {
				return nil
			}
		}
	}
}

func (x *D) readFirmwareStatus(block int) ([]byte, error) {
	req := modbus.Req{
		Addr:     modbus.Addr(block) + 101,
		ProtoCmd: 0x45,
	}

	responseReader := comport.Comm{
		Port:   x.portMeasurer,
		Config: x.cfg.Predefined().ComportMeasurer,
	}

	resp, err := responseReader.GetResponse(req.Bytes())
	if err != nil {
		return nil, err
	}
	if err = req.CheckResponse(resp); err != nil {
		return nil, x.portMeasurer.WrapError(err)
	}
	if len(resp) != 12 {
		return nil, x.portMeasurer.Errorf("ожидалось 12 байт ответа, получено %d", len(resp))
	}
	return resp[2:10], nil
}

func (x *D) writePreparedDataToFlash(block int, placesMask byte, addr uint16, count int) error {
	logrus.Infof("отправка команды записи ранее переданного куска прошивки, %d байт, адрес % X", addr, count)
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

	responseReader := comport.Comm{
		Port:   x.portMeasurer,
		Config: x.cfg.Predefined().ComportMeasurer,
	}
	resp, err := responseReader.GetResponse(req.Bytes())
	if err != nil {
		return err
	}
	if err = req.CheckResponse(resp); err != nil {
		return x.portMeasurer.WrapError(err)
	}

	if !compareBytes(resp, req.Bytes()) {
		return x.portMeasurer.Errorf("% X != % X", resp, req.Bytes())
	}

	return nil
}

func (x *D) sendDataToWriteFlash(block, placeInBlock int, b []byte) error {
	logrus.Infof("отправка куска прошивки для записи: %d байт, % X", len(b), b)

	req := modbus.Req{
		Addr:     modbus.Addr(block) + 101,
		ProtoCmd: 0x42,
		Data: append([]byte{
			byte(placeInBlock + 1),
			byte(len(b) >> 8),
			byte(len(b)),
		}, b...),
	}
	responseReader := comport.Comm{
		Port:   x.portMeasurer,
		Config: x.cfg.Predefined().ComportMeasurer,
	}
	resp, err := responseReader.GetResponse(req.Bytes())

	if err != nil {
		return err
	}
	if err = req.CheckResponse(resp); err != nil {
		return x.portMeasurer.WrapError(err)
	}
	if len(resp) != 7 {
		return x.portMeasurer.Errorf("длина ответа %d не равна 7", len(resp))
	}
	if !compareBytes(resp[:5], req.Bytes()[:5]) {
		return x.portMeasurer.Errorf("% X != % X", resp[:5], req.Bytes()[:5])
	}
	return nil
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
