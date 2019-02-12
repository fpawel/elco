package daemon

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/goutils/serial-comm/comport"
	"github.com/fpawel/goutils/serial-comm/modbus"
	"github.com/hako/durafmt"
	"github.com/sirupsen/logrus"
	"time"
)

func (x *D) writePartyFirmware() error {

	c := x.c.LastParty()
	blockProducts := GroupProductsByBlocks(c.ProductsWithProduction())

	logrus.WithFields(logrus.Fields{
		"party_id": c.Party().PartyID,
	}).Info("write party firmware")

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

	strProducts := ""

	for i, p := range products {
		if i > 0 {
			strProducts += ", "
		}
		strProducts += fmt.Sprintf("%d", p.Place%8)
	}

	x.hardware.logFields["block"] = block
	x.hardware.logFields["products"] = strProducts
	x.hardware.logFields["places_mask"] = fmt.Sprintf("%08b", placesMask)
	defer func() {
		delete(x.hardware.logFields, "block")
		delete(x.hardware.logFields, "products")
		delete(x.hardware.logFields, "places_mask")

	}()

	logrus.Info("write products firmware")

	firmwareBytes := make(map[int]data.FirmwareBytes)

	for _, p := range products {
		pi := x.c.PartiesCatalogue().ProductInfo(p.ProductID)
		firmware, err := pi.Firmware()
		if err != nil {
			return merry.Appendf(err, "расчёт не удался для места %d.%d",
				p.Place/8+1, p.Place%8+1)
		}
		firmwareBytes[p.Place%8] = firmware.Bytes()
	}

	for _, c := range firmwareAddresses {

		logrus.WithFields(logrus.Fields{
			"addr1": c.addr1,
			"addr2": c.addr2,
		}).Info("write batch")

		for _, p := range products {
			place := p.Place % 8
			d := firmwareBytes[place]
			if err := x.sendDataToWriteFlash(block, place, d[c.addr1:c.addr2+1]); err != nil {
				return err
			}
		}
		if err := x.writePreparedDataToFlash(block, placesMask, c.addr1, int(c.addr2-c.addr1+1)); err != nil {
			return err
		}

		if err := x.waitFirmwareStatus(block, placesMask); err != nil {
			return err
		}
	}
	return nil
}

func (x *D) writeSingleProductFirmware(number int, bytes []byte) error {
	block := number / 8
	place := number % 8
	logrus.WithFields(logrus.Fields{
		"block": block,
		"place": place,
		"bytes": fmt.Sprintf("% X", bytes),
	}).Info("write single product firmware")

	placesMask := byte(1) << byte(place)

	for _, c := range firmwareAddresses {
		logrus.WithFields(logrus.Fields{
			"block":       block,
			"places_mask": fmt.Sprintf("%08b", placesMask),
			"addr1":       c.addr1,
			"addr2":       c.addr2,
			"bytes":       fmt.Sprintf("% X", bytes[c.addr1:c.addr2+1]),
		}).Info("write single product firmware batch")

		if err := x.sendDataToWriteFlash(block, place, bytes[c.addr1:c.addr2+1]); err != nil {
			return err
		}

		if err := x.writePreparedDataToFlash(block, placesMask, c.addr1, int(c.addr2-c.addr1+1)); err != nil {
			return err
		}

		if err := x.waitFirmwareStatus(block, placesMask); err != nil {
			return err
		}
	}
	logrus.WithFields(logrus.Fields{
		"block": block,
		"place": place,
		"bytes": fmt.Sprintf("% X", bytes),
	}).Info("write single product firmware: ok")
	return nil
}

func (x *D) waitFirmwareStatus(block int, placesMask byte) error {

	t := time.Duration(x.sets.Config().Predefined.FirmwareWriter.StatusTimeoutSeconds) * time.Second
	ctx, _ := context.WithTimeout(x.hardware.ctx, t)

	for {

		select {
		case <-ctx.Done():
			status, err := x.readFirmwareStatus(block)
			if err != nil {
				return x.port.measurer.LastWork().WrapError(err)
			}
			if err = checkFirmwareStatus(status, placesMask); err != nil {
				err = merry.Wrap(err).WithValue("status_timeout", durafmt.Parse(t))
				return err
			}
			return nil

		default:
			status, err := x.readFirmwareStatus(block)
			if err != nil {
				return x.port.measurer.LastWork().WrapError(err)
			}
			if err = checkFirmwareStatus(status, placesMask); err == nil {
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
		Port:   x.port.measurer,
		Config: x.sets.Config().MeasurerComm,
	}

	resp, err := responseReader.GetResponse(req.Bytes())
	if err != nil {
		return nil, err
	}
	w := x.port.measurer.LastWork()
	if err = req.CheckResponse(resp); err != nil {
		return nil, w.WrapError(err)
	}
	if len(resp) != 12 {
		return nil, w.Errorf("ожидалось 12 байт ответа, получено %d", len(resp))
	}
	return resp[2:], nil
}

func (x *D) writePreparedDataToFlash(block int, placesMask byte, addr uint16, count int) error {
	logrus.WithFields(logrus.Fields{
		"block":       block,
		"places_mask": fmt.Sprintf("%08b", placesMask),
		"addr":        addr,
		"count":       count,
	}).Info()
	req := modbus.Req{
		Addr:     modbus.Addr(block) + 101,
		ProtoCmd: 0x43,
		Data: []byte{
			placesMask,
			byte(x.sets.Config().UserConfig.Firmware.ChipType),
			byte(addr >> 8),
			byte(addr),
			byte(count >> 8),
			byte(count),
		},
	}

	responseReader := comport.Comm{
		Port:   x.port.measurer,
		Config: x.sets.Config().MeasurerComm,
	}
	resp, err := responseReader.GetResponse(req.Bytes())
	if err != nil {
		return err
	}
	if !compareBytes(resp, req.Bytes()) {
		return x.port.measurer.LastWork().Errorf("% X != % X", resp, req.Bytes())
	}

	return nil
}

func checkFirmwareStatus(b []byte, placesMask byte) error {
	for i := byte(0); i < 8; i++ {
		if (1<<i)&placesMask != 0 && b[i] != 0 {
			return merry.New("не верный код статуса").
				WithValue("place", i).
				WithValue("status", fmt.Sprintf("%X", b[i]))
		}
	}
	return nil
}

func (x *D) sendDataToWriteFlash(block, place int, b []byte) error {
	logrus.WithFields(logrus.Fields{
		"block":    block,
		"place":    place,
		"data":     fmt.Sprintf("% X", b),
		"data_len": len(b),
	}).Info()

	req := modbus.Req{
		Addr:     modbus.Addr(block) + 101,
		ProtoCmd: 0x42,
		Data: append([]byte{
			byte(place + 1),
			byte(len(b) >> 8),
			byte(len(b)),
		}, b...),
	}
	responseReader := comport.Comm{
		Port:   x.port.measurer,
		Config: x.sets.Config().MeasurerComm,
	}
	resp, err := responseReader.GetResponse(req.Bytes())

	if err != nil {
		return err
	}
	w := x.port.measurer.LastWork()
	if err = req.CheckResponse(resp); err != nil {
		return w.WrapError(err)
	}
	if len(resp) != 7 {
		return w.Errorf("длина ответа %d не равна 7", len(resp))
	}
	if !compareBytes(resp[:5], req.Bytes()[:5]) {
		return w.Errorf("% X != % X", resp[:5], req.Bytes()[:5])
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

func merryWithValues(e error, values logrus.Fields) merry.Error {
	err := merry.Wrap(e)
	for k, v := range values {
		err = err.WithValue(k, v)
	}
	return err
}

var firmwareAddresses = []struct{ addr1, addr2 uint16 }{
	{0, 512},
	{1024, 1535},
	{1536, 1600},
	{1792, 1810},
	{1824, 1831},
}
