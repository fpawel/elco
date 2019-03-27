package modbus

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/pkg/errfmt"
	"github.com/fpawel/elco/pkg/serial-comm/comm"
	"github.com/hashicorp/go-multierror"
)

type ResponseReader interface {
	GetResponse([]byte, comm.ResponseParser) ([]byte, error)
}

func read3(responseReader ResponseReader, addr Addr, firstReg Var, regsCount uint16,
	parseResponse comm.ResponseParser) ([]byte, error) {
	req := Req{
		Addr:     addr,
		ProtoCmd: 3,
		Data:     append(uint16b(uint16(firstReg)), uint16b(regsCount)...),
	}
	request := req.Bytes()
	response, err := responseReader.GetResponse(request, func(request, response []byte) error {

		if err := req.CheckResponse(response); err != nil {
			return err
		}
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
		return response, errfmt.WithReqResp(err, request, response)
	}
	return response, nil
}

func Read3BCDValues(responseReader ResponseReader, addr Addr, var3 Var, count int) ([]float64, error) {
	var values []float64
	_, err := read3(responseReader, addr, var3, uint16(count*2),
		func(request, response []byte) error {
			var err error
			for i := 0; i < count; i++ {
				n := 3 + i*4
				if v, ok := ParseBCD6(response[n:]); !ok {
					err = multierror.Append(err,
						fmt.Errorf("не правильный код BCD: n=%d BCD=%X", n, response[n:n+4]))
				} else {
					values = append(values, v)
				}
			}
			if err != nil {
				return merry.WithMessagef(err, "addr=%d var3=%d count=%d: %s", addr, var3, count)
			}
			return nil
		})
	return values, err

}

func Read3BCD(responseReader ResponseReader, addr Addr, var3 Var) (result float64, err error) {

	_, err = read3(responseReader, addr, var3, 2,
		func(request []byte, response []byte) error {
			var ok bool
			if result, ok = ParseBCD6(response[3:]); !ok {
				return merry.Errorf("не правильный код BCD: % X", response[3:7])
			}
			return nil
		})
	return
}

//func Write32FloatProto(r responseGetter, addr Addr, protocolCommandCode ProtoCmd,
//	deviceCommandCode DevCmd, value float64) error {
//	req := Write32BCDRequest(addr, protocolCommandCode, deviceCommandCode, value)
//	response, err := r.GetResponse(req.Bytes())
//
//	if err == nil {
//		err = req.CheckResponse16(response)
//	}
//
//	if err != nil {
//		err =  merry.Wrap(err).
//			WithValue( "request", req.Bytes()).
//			WithValue("response", response).
//			WithMessagef("не удалась команда %d(%X) с аргументом %v", deviceCommandCode, deviceCommandCode, value)
//	}
//	return err
//}
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
