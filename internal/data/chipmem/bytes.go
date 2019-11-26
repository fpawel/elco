package chipmem

import (
	"encoding/binary"
	"github.com/fpawel/elco/internal/data"
	"math"
	"time"
)

type Bytes []byte

//func FirmwareBytes(s data.Product) (b Bytes, err error) {
//	if len(s.Firmware) == 0 {
//		err = merry.New("микросхема ЭХЯ не запрогромирована")
//		return
//	}
//	if len(s.Firmware) < FirmwareSize {
//		err = merry.New("не верный формат данных в памяти микросхемы")
//		return
//	}
//	b = s.Firmware
//	return
//}

func (x Bytes) Time() time.Time {
	_ = x[0x0712]
	return time.Date(
		2000+int(x[0x070F]),
		time.Month(x[0x070E]),
		int(x[0x070D]),
		int(x[0x0710]),
		int(x[0x0711]),
		int(x[0x0712]), 0, time.UTC)
}

func (x Bytes) ProductType() string {
	const offset = 0x060B
	n := offset
	for ; n < offset+50; n++ {
		if x[n] == 0xff || x[n] == 0 {
			break
		}
	}
	return string(x[offset:n])
}

func (x Bytes) FirmwareInfo() FirmwareInfo {
	t := x.Time()
	r := FirmwareInfo{
		ProductTempPoints:  x.ProductTempPoints(),
		Year:               t.Year(),
		Month:              int(t.Month()),
		Day:                t.Day(),
		Hour:               t.Hour(),
		Minute:             t.Minute(),
		Second:             t.Second(),
		ProductType:        x.ProductType(),
		Serial:             formatBCD(x[0x0701:0x0705], -1),
		ScaleBeg:           formatBCD(x[0x0602:0x0606], -1),
		ScaleEnd:           formatBCD(x[0x0606:0x060A], -1),
		SensitivityLab73:   formatFloat(math.Float64frombits(binary.LittleEndian.Uint64(x[0x0720:])), 3),
		SensitivityProduct: formatBCD(x[0x0709:0x070D], -1),
		Fon20:              formatBCD(x[0x0705:0x0709], -1),
		TempValues:         TempValues(x.ProductTempPoints()),
		Units:              unitsName(x[0x060A]),
		Gas:                gasName(x[0x0600]),
	}
	return r
}

func (x Bytes) ProductTempPoints() (r TempPoints) {

	valAt := func(i int) float64 {
		a := binary.LittleEndian.Uint16(x[i:])
		b := int16(a)
		y := float64(b)
		return y
	}

	t := float64(-124)
	n := 0
	for i := 0x00F8; i >= 0; i -= 2 {
		r.Temp[n] = t
		r.Fon[n] = valAt(i)
		t++
		n++
	}
	t = 0
	for i := 0x0100; i <= 0x01F8; i += 2 {
		r.Temp[n] = t
		r.Fon[n] = valAt(i)
		t++
		n++
	}
	t = -124
	n = 0
	for i := 0x04F8; i >= 0x0400; i -= 2 {
		r.Sens[n] = valAt(i)
		t++
		n++
	}
	t = 0
	for i := 0x0500; i <= 0x05F8; i += 2 {
		r.Sens[n] = valAt(i)
		t++
		n++
	}
	return
}

func unitsName(code byte) string {
	for _, a := range data.ListUnits() {
		if a.Code == code {
			return a.UnitsName
		}
	}
	return ""
}

func gasName(code byte) string {
	for _, a := range data.ListGases() {
		if a.Code == code {
			return a.GasName
		}
	}
	return ""
}
