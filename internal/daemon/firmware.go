package daemon

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/pkg/errfmt"
	"github.com/fpawel/elco/pkg/serial-comm/comport"
	"github.com/fpawel/elco/pkg/serial-comm/modbus"
	"github.com/hako/durafmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/reform.v1"
	"sort"
	"strings"
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
	})
	if err != nil {
		return err
	}

	var placesStr []string
	for _, p := range products {
		placesStr = append(placesStr, data.FormatPlace(p.Place))
	}
	sort.Strings(placesStr)
	m := logrus.Fields{
		"party_id": party.PartyID,
		"places":   strings.Join(placesStr, ", "),
	}

	//logrus.WithFields(m).Info("Начало записи прошивки партии", )

	placeBytes := map[int][]byte{}

	blockProducts := GroupProductsByBlocks(products)
	for _, products := range blockProducts {
		if err := x.writeProductsFirmware(products, placeBytes); err != nil {
			return err
		}
	}
	m["продолжительность_записи"] = durafmt.Parse(time.Since(startTime))

	for _, products := range blockProducts {
		var places []int
		for _, p := range products {
			places = append(places, p.Place)
		}
		sort.Ints(places)
		if err := x.verifyProductsFirmware(places, placeBytes); err != nil {
			return err
		}
	}

	if err = data.GetPartyProducts(x.dbProducts, &party); err != nil {
		return err
	}
	notify.LastPartyChanged(x.w, party)

	logrus.WithFields(m).Info("Запись прошивки партии завершена")

	return nil
}

func (x *D) verifyProductsFirmware(places []int, placeBytes map[int][]byte) error {
	for _, place := range places {
		b, err := x.readFirmware(place)
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

func (x *D) writeProductsFirmware(products []*data.Product, placeBytes map[int][]byte) error {

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

	//logrus.Info("прошивка блока", )

	for _, p := range products {
		prodInfo := new(data.ProductInfo)
		if err := x.dbProducts.FindByPrimaryKeyTo(prodInfo, p.ProductID); err != nil {
			return err
		}
		firmware, err := prodInfo.Firmware()
		if err != nil {
			return merry.Appendf(err, "расчёт прошивки ЭХЯ не удался %v", prodInfo)
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

		if err := x.sendDataToWriteFlash(block, placeInBlock, b[addr1:addr2+1]); err != nil {
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
		if err := x.waitFirmwareStatus(block, placesMask); err != nil {
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
			return nil, err
		}
		if len(resp) != 10+int(count) {
			return nil, errfmt.WithReqRespMsgf(req.Bytes(), resp, "ожидалось %d байт ответа, получено %d",
				10+int(count), len(resp))

		}
		copy(b[c.addr1:c.addr1+count], resp[8:8+count])
		if i < len(firmwareAddresses)-1 {
			time.Sleep(time.Duration(x.cfg.Predefined().ReadRangeDelayMillis) * time.Millisecond)
		}
	}
	//logrus.Infof("считана прошивка ЭХЯ: %d байт, % X", len(b), b)
	return b, nil
}

func (x *D) writeFirmware(place int, bytes []byte) error {
	x.logFields["place"] = place
	defer delete(x.logFields, "place")
	//logrus.Infof("запись прошивки ЭХЯ: %d байт, % X", len(bytes), bytes)

	block := place / 8
	placeInBlock := place % 8
	placesMask := byte(1) << byte(placeInBlock)

	doAddresses := func(addr1, addr2 uint16) error {
		x.logFields["адрес_начала_куска"] = addr1
		x.logFields["адрес_конца_куска"] = addr2
		x.logFields["количество_байт_куска"] = addr2 + 1 - addr1
		defer delete(x.logFields, "адрес_начала_куска")
		defer delete(x.logFields, "адрес_конца_куска")
		defer delete(x.logFields, "количество_байт_куска")

		//logrus.WithFields(logrus.Fields{}).Infof("запись куска прошивки ЭХЯ: %d...%d, %d байт", addr1, addr2, addr2+1-addr1)

		if err := x.sendDataToWriteFlash(block, placeInBlock, bytes[addr1:addr2+1]); err != nil {
			return err
		}

		if err := x.writePreparedDataToFlash(block, placesMask, addr1, int(addr2-addr1+1)); err != nil {
			return err
		}
		time.Sleep(time.Second)

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
	return nil
}

func (x *D) waitFirmwareStatus(block int, placesMask byte) error {
	//startTime := time.Now()
	//defer func() {
	//	logrus.Infof("ожидание статуса: блок %d, %08b: %s",
	//		block, placesMask, durafmt.Parse(time.Since(startTime)))
	//}()

	t := time.Duration(x.cfg.Predefined().StatusTimeoutSeconds) * time.Second
	//logrus.Infof("прошивка блока %d: ожидание статуса завершения, таймаут %s", block, durafmt.Parse(t))
	ctx, _ := context.WithTimeout(x.hardware.ctx, t)
	for {

		select {
		case <-ctx.Done():
			request, response, err := x.readFirmwareStatus(block)
			if err != nil {
				return err
			}
			status := response[2:10]

			for i, b := range status {
				if (1<<byte(i))&placesMask != 0 && b != 0 {
					return errfmt.WithReqRespMsgf(request, response,
						"не удалось записать прошивку места %d: таймаут %s, статус[%d]=%X",
						durafmt.Parse(t), block*8+i+1, i, b)
				}
			}
			return nil

		default:
			request, response, err := x.readFirmwareStatus(block)
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
						return errfmt.WithReqRespMsgf(request, response, "не удалось записать прошивку места %d: статус[%d]=%X",
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

func (x *D) readFirmwareStatus(block int) (request []byte, response []byte, err error) {
	req := modbus.Req{
		Addr:     modbus.Addr(block) + 101,
		ProtoCmd: 0x45,
	}
	request = req.Bytes()

	responseReader := comport.Comm{
		Port:   x.portMeasurer,
		Config: x.cfg.Predefined().ComportMeasurer,
	}

	response, err = responseReader.GetResponse(req.Bytes())
	if err != nil {
		return
	}
	if err = req.CheckResponse(response); err != nil {
		return
	}
	if len(response) != 12 {
		err = errfmt.WithReqRespMsgf(request, response, "ожидалось 12 байт ответа, получено %d", len(response))
		return
	}
	return
}

func (x *D) writePreparedDataToFlash(block int, placesMask byte, addr uint16, count int) error {
	//logrus.Infof("отправка команды записи ранее переданного куска прошивки, %d байт, адрес % X", addr, count)
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

	responseReader := comport.Comm{
		Port:   x.portMeasurer,
		Config: x.cfg.Predefined().ComportMeasurer,
	}
	response, err := responseReader.GetResponse(request)
	if err != nil {
		return err
	}
	if err = req.CheckResponse(response); err != nil {
		return err
	}

	if !compareBytes(response, req.Bytes()) {
		return errfmt.WithReqRespMsg(request, response, "запрос не равен ответу")
	}

	return nil
}

func (x *D) sendDataToWriteFlash(block, placeInBlock int, b []byte) error {
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
	responseReader := comport.Comm{
		Port:   x.portMeasurer,
		Config: x.cfg.Predefined().ComportMeasurer,
	}
	response, err := responseReader.GetResponse(request)

	if err != nil {
		return err
	}
	if err = req.CheckResponse(response); err != nil {
		return err
	}
	if len(response) != 7 {
		return errfmt.WithReqRespMsgf(request, response, "длина ответа %d не равна 7", len(response))
	}
	if !compareBytes(response[:5], req.Bytes()[:5]) {
		return errfmt.WithReqRespMsgf(request, response, "% X != % X", response[:5], request[:5])
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
