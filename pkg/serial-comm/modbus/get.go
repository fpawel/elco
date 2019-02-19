package modbus

import (
	"fmt"
	"github.com/fpawel/elco/pkg/errfmt"
	"github.com/hashicorp/go-multierror"
)

type responseGetter interface {
	GetResponse([]byte) ([]byte, error)
}

func read3(responseReader responseGetter, addr Addr, firstReg Var, regsCount uint16) ([]byte, []byte, error) {
	req := Req{
		Addr:     addr,
		ProtoCmd: 3,
		Data:     append(uint16b(uint16(firstReg)), uint16b(regsCount)...),
	}
	request := req.Bytes()
	response, err := responseReader.GetResponse(request)
	if err != nil {
		return request, nil, err
	}
	err = req.CheckResponse(response)
	if err != nil {
		return request, nil, err
	}
	lenMustBe := int(regsCount)*2 + 5
	if len(response) != lenMustBe {
		return request, nil, errfmt.WithReqResp(ProtocolError.Here(), request, response).
			WithMessagef("длина ответа %d не равна %d", len(response), lenMustBe)
	}
	return request, response, nil
}

type ReadBCDValuesResult struct {
	Request, Response []byte
	Values            []float64
}

func Read3BCDValues(responseReader responseGetter, addr Addr, var3 Var, count int) (r ReadBCDValuesResult, err error) {
	r.Request, r.Response, err = read3(responseReader, addr, var3, uint16(count*2))
	if err != nil {
		return
	}
	for i := 0; i < count; i++ {
		n := 3 + i*4
		if v, ok := ParseBCD6(r.Response[n:]); !ok {
			err = multierror.Append(err,
				fmt.Errorf("не правильный код BCD: n=%d BCD=%X", n, r.Response[n:n+4]))
		} else {
			r.Values = append(r.Values, v)
		}
	}
	if err != nil {
		err = errfmt.WithReqResp(ProtocolError.Here(), r.Request, r.Response).
			WithMessagef("addr=%d var3=%d count=%d: %s", addr, var3, count, err.Error())
	}
	return
}

//func Read3(responseReader responseGetter, addr Addr, firstReg Var, regsCount uint16) ([]byte, error) {
//	_, response, err := read3(responseReader, addr, firstReg, regsCount)
//	return response, err
//}
//func Read3BCD(responseReader responseGetter, addr Addr, var3 Var) (float64, error) {
//	request, response, err := read3(responseReader, addr, var3, 2)
//	if err != nil {
//		return 0, err
//	}
//	if v, ok := ParseBCD6(response[3:]); !ok {
//		return 0, ProtocolError.Here().
//			WithValue( "request", request).
//			WithValue("response", response).
//			WithMessagef("не правильный код BCD: % X", response[3:7])
//	} else {
//		return v, nil
//	}
//}
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
