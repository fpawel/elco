package modbus

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/pkg/serial-comm/comm"
)

var ErrProtocol = merry.WithMessage(comm.ErrProtocol, "modbus error")

func (x Req) CheckResponse(response []byte) error {
	return x.checkResponse(response)
}

func (x Req) checkResponse(response []byte) error {

	if len(response) == 0 {
		return ErrProtocol.Here().WithMessage("нет ответа")
	}

	if len(response) < 4 {
		return ErrProtocol.Here().WithMessage("длина ответа меньше 4")
	}

	if h, l := CRC16(response); h != 0 || l != 0 {
		return ErrProtocol.Here().WithMessage("CRC16 не ноль")
	}
	if response[0] != byte(x.Addr) {
		return ErrProtocol.Here().WithMessagef("несовпадение адресов запроса [%d] и ответа [%d]",
			x.Addr, response[0])
	}

	if len(response) == 5 && byte(x.ProtoCmd)|0x80 == response[1] {
		return ErrProtocol.Here().WithMessagef("код ошибки %X", response[2])
	}
	if response[1] != byte(x.ProtoCmd) {
		return ErrProtocol.Here().WithMessagef("несовпадение кодов команд запроса [%d] и ответа [%d]",
			x.ProtoCmd, response[1])
	}

	return nil
}
