package modbus

type ProtoCmd byte
type Addr byte

type Var uint16

type Req struct {
	Addr     Addr
	ProtoCmd ProtoCmd
	Data     []byte
}

type DevCmd uint16

type Coefficient uint16

func (x Req) Bytes() (b []byte) {
	b = make([]byte, 4+len(x.Data))
	b[0] = byte(x.Addr)
	b[1] = byte(x.ProtoCmd)
	copy(b[2:], x.Data)
	n := 2 + len(x.Data)
	b[n], b[n+1] = CRC16(b[:n])
	return
}

func Write32BCDRequest(addr Addr, protocolCommandCode ProtoCmd, deviceCommandCode DevCmd,
	value float64) Req {
	r := Req{
		Addr:     addr,
		ProtoCmd: protocolCommandCode,
	}
	r.Data = []byte{0, 32, 0, 3, 6}
	r.Data = append(r.Data, uint16b(uint16(deviceCommandCode))...)
	r.Data = append(r.Data, BCD6(value)...)
	return r
}

func NewSwitchGasOven(n byte) Req {
	return Req{
		Addr:     0x16,
		ProtoCmd: 0x10,
		Data:     []byte{0, 32, 0, 1, 2, 0, n},
	}
}

func (x *Req) ParseBCDValue(b []byte) (v float64, err error) {
	if err = x.CheckResponse(b); err != nil {
		return
	}
	if len(b) != 9 {
		err = ProtocolError.Here().WithMessagef("длина ответа %d не равна 9", len(b))
		return
	}
	var ok bool
	if v, ok = ParseBCD6(b[3:]); !ok {
		err = ProtocolError.Here().WithMessagef("не правильный код BCD [% X]", b[3:7])
	}
	return

}

func uint16b(v uint16) (b []byte) {
	b = make([]byte, 2)
	b[0] = byte(v >> 8)
	b[1] = byte(v)
	return
}
