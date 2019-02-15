package modbus

import (
	"github.com/ansel1/merry"
)

var ProtocolError = merry.New("modbus")

func (x Req) CheckResponse(b []byte) error {

	if len(b) < 4 {
		return ProtocolError.Here().WithMessage("длина ответа меньше 4")
	}

	if h, l := CRC16(b); h != 0 || l != 0 {
		return ProtocolError.Here().WithMessage("CRC16 не ноль")
	}
	if b[0] != byte(x.Addr) {
		return ProtocolError.Here().WithMessagef("несовпадение адресов запроса [%d] и ответа [%d]",
			x.Addr, b[0])
	}

	if len(b) == 5 && byte(x.ProtoCmd)|0x80 == b[1] {
		return ProtocolError.Here().WithMessagef("прибор вернул код ошибки %d", b[2])
	}
	if b[1] != byte(x.ProtoCmd) {
		return ProtocolError.Here().WithMessagef("несовпадение кодов команд запроса [%d] и ответа [%d]",
			x.ProtoCmd, b[1])
	}

	return nil
}

func (x Req) CheckResponse16(b []byte) error {
	if err := x.CheckResponse(b); err != nil {
		return err
	}
	a := x.Bytes()
	for i := 2; i < 6; i++ {
		if a[i] != b[i] {
			return ProtocolError.Here().WithMessagef("ошибка формата ответа: [% X] != [% X]", a[2:6], b[2:6])
		}
	}
	return nil
}
