package daemon

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/firmware"
	"github.com/fpawel/goutils/serial-comm/modbus"
	"github.com/sirupsen/logrus"
)

func (x *D) writeFlash() error {

	xs := GroupProductsByBlocks(x.c.LastParty().ProductsWithProduction())
	gases := x.c.ListGases()
	units := x.c.ListUnits()

	for _, products := range xs {

		m := make(map[int][firmware.Size]byte)
		for _, p := range products {
			pi := x.c.PartiesCatalogue().ProductInfo(p.ProductID)
			b, err := firmware.FromProductInfo(pi, gases, units)
			if err != nil {
				return err
			}
			m[p.Place%8] = b.B
		}

		block := products[0].Place / 8

		logrus.WithField("block", block).Info("write flash")

		for _, c := range []struct{ a, b uint16 }{
			{0, 512}, {1024, 1535},
			{1536, 1600}, {1792, 1810}, {1824, 1831},
		} {

			var placesMask byte
			for _, p := range products {
				place := byte(p.Place) % 8
				placesMask |= 1 << place
			}

			logrus.WithFields(logrus.Fields{
				"block":       block,
				"addr1":       c.a,
				"addr2":       c.b,
				"places_mask": fmt.Sprintf("%08b", placesMask),
			}).Info("iterate addresses ranges")

			for _, p := range products {
				place := p.Place % 8
				d := m[place]
				if err := x.sendDataToWriteFlash(block, place, d[c.a:c.b+1]); err != nil {
					return err
				}
			}
			if err := x.writePreparedDataToFlash(block, placesMask, c.a, int(c.b-c.a+1)); err != nil {
				return err
			}
		}

	}

	return nil
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

	resp, err := x.port.measurer.GetResponse(req.Bytes(), x.sets.Config().Measurer)
	if err != nil {
		return err
	}

	if !compareBytes(resp, req.Bytes()) {
		return modbus.ProtocolError.Here().WithMessagef("% X != % X", resp, req.Bytes())
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
			byte(place),
			byte(len(b) >> 8),
			byte(len(b)),
		}, b...),
	}
	resp, err := x.port.measurer.GetResponse(req.Bytes(), x.sets.Config().Measurer)

	if err != nil {
		return err
	}
	if err = req.CheckResponse(b); err != nil {
		return merry.Append(err, x.port.measurer.LastWork().FormatResponse())
	}
	if len(b) != 7 {
		return modbus.ProtocolError.Here().WithMessagef("длина ответа %d не равна 7", len(b))
	}
	if !compareBytes(resp[:5], b[:5]) {
		return modbus.ProtocolError.Here().WithMessagef("% X != % X", resp[:5], b[:5])
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
