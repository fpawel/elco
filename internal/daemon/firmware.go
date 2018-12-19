package daemon

import (
	"github.com/fpawel/goutils/serial-comm/modbus"
)

func (x *D) writePreparedDataToFlash(block int, placesMask byte, flashAddr uint16, count int) error {
	//req := modbus.Req{
	//	Addr:     modbus.Addr(block) + 101,
	//	ProtoCmd: 0x43,
	//	Data: []byte{
	//		placesMask,
	//		byte(flashAddr >> 8),
	//		byte(flashAddr),
	//		byte(count >> 8),
	//		byte(count),
	//	},
	//}
	return nil
}

func (x *D) sendDataToWriteFlash(block, place int, b []byte) error {
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
		return err
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
