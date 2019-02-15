package modbus

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/hashicorp/go-multierror"
)

type responseGetter interface {
	GetResponse([]byte) ([]byte, error)
}

func Read3(responseReader responseGetter, addr Addr, firstReg Var, regsCount uint16) ([]byte, error) {
	req := Req{
		Addr:     addr,
		ProtoCmd: 3,
		Data:     append(uint16b(uint16(firstReg)), uint16b(regsCount)...),
	}

	b, err := responseReader.GetResponse(req.Bytes())
	if err != nil {
		return nil, err
	}
	err = req.CheckResponse(b)
	if err != nil {
		return nil, err
	}
	lenMustBe := int(regsCount)*2 + 5
	if len(b) != lenMustBe {

		return nil, ProtocolError.Here().
			WithValue("addr", addr).
			WithValue("first_reg", firstReg).
			WithValue("regs_count", regsCount).
			WithMessagef("длина ответа %d не равна %d", len(b), lenMustBe)
	}
	return b, nil
}

func Read3BCDValues(responseReader responseGetter, addr Addr, var3 Var, count int) (values []float64, err error) {
	b, err := Read3(responseReader, addr, var3, uint16(count*2))
	if err != nil {
		return nil, err
	}
	for i := 0; i < count; i++ {
		n := 3 + i*4
		if v, ok := ParseBCD6(b[n:]); !ok {
			err = multierror.Append(err,
				fmt.Errorf("не правильный код BCD: addr=%d var3=%d count=%d n=%d BCD=%X",
					addr, var3, count, n, b[n:n+4]))
		} else {
			values = append(values, v)
		}
	}
	if err != nil {
		err = ProtocolError.Here().WithMessage(err.Error())
	}
	return
}

func Read3BCD(responseReader responseGetter, addr Addr, var3 Var) (float64, error) {
	b, err := Read3(responseReader, addr, var3, 2)
	if err != nil {
		return 0, merry.Wrap(err).
			WithValue("addr", addr).
			WithValue("var3", var3)
	}
	if v, ok := ParseBCD6(b[3:]); !ok {
		return 0, ProtocolError.Here().
			WithValue("addr", addr).
			WithValue("var3", var3).
			WithValue("BCD", fmt.Sprintf("% X", b[3:7])).
			WithMessage("не правильный код BCD")
	} else {
		return v, nil
	}
}

func Write32FloatProto(r responseGetter, addr Addr, protocolCommandCode ProtoCmd,
	deviceCommandCode DevCmd, value float64) error {
	req := Write32BCDRequest(addr, protocolCommandCode, deviceCommandCode, value)
	b, err := r.GetResponse(req.Bytes())
	if err != nil {
		return merry.Wrap(err).
			WithValue("addr", addr).
			WithValue("protocol_command_code", protocolCommandCode).
			WithValue("device_command_code", deviceCommandCode).
			WithValue("value", value)
	}
	return req.CheckResponse16(b)
}

func Write32Float(r responseGetter, addr Addr, deviceCommandCode DevCmd, value float64) error {
	return Write32FloatProto(r, addr, 0x10, deviceCommandCode, value)
}

func Write32Float1016(r responseGetter, addr Addr, deviceCommandCode DevCmd, value float64) error {
	err := Write32Float(r, addr, deviceCommandCode, value)
	if err == context.DeadlineExceeded || merry.Is(err, ProtocolError) {
		err = Write32FloatProto(r, addr, 0x16, deviceCommandCode, value)
	}
	return err
}

func ReadCoefficient(r responseGetter, addr Addr, coefficient Coefficient) (float64, error) {
	return Read3BCD(r, addr, 224+2*Var(coefficient))
}

func WriteCoefficientProto(r responseGetter, addr Addr, protocolCommandCode ProtoCmd,
	coefficient Coefficient, value float64) error {
	return Write32FloatProto(r, addr, protocolCommandCode, (0x80<<8)+DevCmd(coefficient), value)
}

func WriteCoefficient(r responseGetter, addr Addr, coefficient Coefficient, value float64) error {
	return WriteCoefficientProto(r, addr, 0x10, coefficient, value)
}

func WriteCoefficient1016(r responseGetter, addr Addr, coefficient Coefficient, value float64) error {
	err := WriteCoefficient(r, addr, coefficient, value)
	if err == context.DeadlineExceeded || merry.Is(err, ProtocolError) {
		err = WriteCoefficientProto(r, addr, 0x16, coefficient, value)
	}
	return err
}
