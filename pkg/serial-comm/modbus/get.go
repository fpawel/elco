package modbus

import (
	"encoding/binary"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/pkg/serial-comm/comm"
	"github.com/hashicorp/go-multierror"
)

func Read3(responseReader ResponseReader, addr Addr, firstReg Var, regsCount uint16, parseResponse comm.ResponseParser) ([]byte, error) {
	req := Req{
		Addr:     addr,
		ProtoCmd: 3,
		Data:     append(uint16b(uint16(firstReg)), uint16b(regsCount)...),
	}

	response, err := req.GetResponse(responseReader, func(request, response []byte) error {
		lenMustBe := int(regsCount)*2 + 5
		if len(response) != lenMustBe {
			return merry.Errorf("длина ответа %d не равна %d", len(response), lenMustBe)
		}
		if parseResponse != nil {
			return parseResponse(request, response)
		}
		return nil
	})

	if err != nil {
		err = merry.Appendf(err, "чтение регистр=%d количество_регистров=%d", firstReg, regsCount)
	}

	return response, err
}

func Read3BCDValues(responseReader ResponseReader, addr Addr, var3 Var, count int) ([]float64, error) {
	var values []float64
	_, err := Read3(responseReader, addr, var3, uint16(count*2),
		func(request, response []byte) error {
			var err error
			for i := 0; i < count; i++ {
				n := 3 + i*4
				if v, ok := ParseBCD6(response[n:]); !ok {
					err = multierror.Append(err,
						fmt.Errorf("не правильный код BCD: позиция=%d BCD=%X", n, response[n:n+4]))
				} else {
					values = append(values, v)
				}
			}
			return err
		})
	if err != nil {
		err = merry.Appendf(err, "запрос %d значений BCD", count)
	}
	return values, err

}

func Read3BCD(responseReader ResponseReader, addr Addr, var3 Var) (result float64, err error) {

	_, err = Read3(responseReader, addr, var3, 2,
		func(request []byte, response []byte) error {
			var ok bool
			if result, ok = ParseBCD6(response[3:]); !ok {
				return merry.Errorf("не правильный код BCD: % X", response[3:7])
			}
			return nil
		})
	if err != nil {
		err = merry.Append(err, "запрос значения BCD")
	}
	return
}

func Write32FloatProto(responseReader ResponseReader, addr Addr, protocolCommandCode ProtoCmd,
	deviceCommandCode DevCmd, value float64) error {
	req := Write32BCDRequest(addr, protocolCommandCode, deviceCommandCode, value)

	_, err := req.GetResponse(responseReader, func(request, response []byte) error {
		for i := 2; i < 6; i++ {
			if request[i] != response[i] {
				return ErrProtocol.Here().
					WithMessagef("ошибка формата ответа: [% X] != [% X]", request[2:6], response[2:6])
			}
		}
		return nil
	})

	if err != nil {
		err = merry.Appendf(err, "запись регистра 32 cmd=%d arg=%v", deviceCommandCode, value)
	}
	return err
}

//func ReadFloat32(responseReader ResponseReader, addr Addr, var3 Var) (result float32, err error) {
//	_, err = Read3(responseReader, addr, var3, 2,
//		func(_, response []byte) error {
//			bits := binary.LittleEndian.Uint32(response[3:7])
//			result = math.Float32frombits(bits)
//			return nil
//		})
//	return
//}

func ReadUInt16(responseReader ResponseReader, addr Addr, var3 Var) (result uint16, err error) {
	_, err = Read3(responseReader, addr, var3, 1,
		func(_, response []byte) error {
			result = binary.LittleEndian.Uint16(response[3:5])
			return nil
		})
	return
}

//func Write32Float(r responseGetter, addr Addr, deviceCommandCode DevCmd, value float64) error {
//	return Write32FloatProto(r, addr, 0x10, deviceCommandCode, value)
//}
//func Write32Float1016(r responseGetter, addr Addr, deviceCommandCode DevCmd, value float64) error {
//	err := Write32Float(r, addr, deviceCommandCode, value)
//	if err == context.DeadlineExceeded || merry.Is(err, ProtocolError) {
//		err = Write32FloatProto(r, addr, 0x16, deviceCommandCode, value)
//	}
//	return err
//}
//
//func ReadCoefficient(r responseGetter, addr Addr, coefficient Coefficient) (float64, error) {
//	return Read3BCD(r, addr, 224+2*Var(coefficient))
//}
//func WriteCoefficientProto(r responseGetter, addr Addr, protocolCommandCode ProtoCmd,
//	coefficient Coefficient, value float64) error {
//	return Write32FloatProto(r, addr, protocolCommandCode, (0x80<<8)+DevCmd(coefficient), value)
//}
//func WriteCoefficient(r responseGetter, addr Addr, coefficient Coefficient, value float64) error {
//	return WriteCoefficientProto(r, addr, 0x10, coefficient, value)
//}
//func WriteCoefficient1016(r responseGetter, addr Addr, coefficient Coefficient, value float64) error {
//	err := WriteCoefficient(r, addr, coefficient, value)
//	if err == context.DeadlineExceeded || merry.Is(err, ProtocolError) {
//		err = WriteCoefficientProto(r, addr, 0x16, coefficient, value)
//	}
//	return err
//}
