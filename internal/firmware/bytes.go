package firmware

import (
	"encoding/binary"
	"fmt"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/goutils"
	"github.com/pkg/errors"
	"math"
	"strconv"
	"time"
)

const Size = 1832

type Bytes struct {
	gases []data.Gas
	units []data.Units
	b     [Size]byte
}

type CodeStr struct {
	Code byte
	Str  string
}

func FromBytes(b []byte, gases []data.Gas, units []data.Units) (*Bytes, error) {
	if len(b) == 0 {
		return nil, errors.New("ЭХЯ не \"прошита\"")
	}
	if len(b) < Size {
		return nil, errors.New("не верный формат \"прошивки\"")
	}
	x := &Bytes{
		gases: gases,
		units: units,
	}
	copy(x.b[:Size], b)
	return x, nil
}

func FromProductInfo(p data.ProductInfo, gases []data.Gas, units []data.Units) (*Bytes, error) {
	x := &Bytes{
		gases: gases,
		units: units,
	}
	if !p.Serial.Valid {
		return x, errors.New("не задан серийный номер")
	}
	if !p.KSens20.Valid {
		return x, errors.New("нет значения к-та чувствительности")
	}
	for i := 0; i < len(x.b); i++ {
		x.b[i] = 0xFF
	}
	goutils.PutBCD6(x.b[0x0701:], float64(p.Serial.Int64))
	goutils.PutBCD6(x.b[0x0602:], 0)
	goutils.PutBCD6(x.b[0x0606:], p.Scale)

	x.b[0x070F] = byte(p.CreatedAt.Year() - 2000)
	x.b[0x070E] = byte(p.CreatedAt.Month())
	x.b[0x070D] = byte(p.CreatedAt.Day())
	x.b[0x0710] = byte(p.CreatedAt.Hour())
	x.b[0x0711] = byte(p.CreatedAt.Minute())
	x.b[0x0712] = byte(p.CreatedAt.Second())
	x.b[0x0600] = p.GasCode
	x.b[0x060A] = p.UnitsCode
	x.PutProductTypeName(p.AppliedProductTypeName)
	binary.LittleEndian.PutUint64(x.b[0x0720:], math.Float64bits(p.KSens20.Float64))
	if err := x.PutProductFonPoints(p); err != nil {
		return x, err
	}
	if err := x.PutProductSensPoints(p); err != nil {
		return x, err
	}
	return x, nil
}

//PutProductTypeName - исполнение
func (x *Bytes) PutProductTypeName(productTypeName string) {
	const n = 50
	b := []byte(productTypeName)
	if len(b) > n {
		b = b[:n]
	}
	for i := copy(x.b[0x060B:], b); i < n; i++ {
		x.b[i] = 0xFF
	}
}

func (x *Bytes) putTempValue(value float64, i int) {
	y := math.Round(value)
	n := uint16(y)
	binary.LittleEndian.PutUint16(x.b[i:], n)
}

//PutProductFonPoints - записать в буфер точки фоновых токов
func (x *Bytes) PutProductFonPoints(p data.ProductInfo) error {
	xy, err := srcFon(p)
	for k := range xy {
		xy[k] *= 1000
	}

	if err != nil {
		return err
	}
	at := newApproxTbl(xy)
	t := float64(-124)
	for i := 0x00F8; i >= 0; i -= 2 {
		x.putTempValue(at.F(t), i)
		t++
	}
	t = 0
	for i := 0x0100; i <= 0x01F8; i += 2 {
		x.putTempValue(at.F(t), i)
		t++
	}
	return nil
}

//PutProductSensPoints - записать в буфер точки коэфф. чувствительности
func (x *Bytes) PutProductSensPoints(p data.ProductInfo) error {
	xy, err := srcSens(p)
	if err != nil {
		return err
	}
	at := newApproxTbl(xy)
	t := float64(-124)
	for i := 0x04F8; i >= 0x0400; i -= 2 {
		x.putTempValue(at.F(t), i)
		t++
	}
	t = 0
	for i := 0x0500; i <= 0x05F8; i += 2 {
		x.putTempValue(at.F(t), i)
		t++
	}
	return nil
}

func (x *Bytes) Time() time.Time {
	_ = x.b[0x0712]
	return time.Date(
		2000+int(x.b[0x070F]),
		time.Month(x.b[0x070E]),
		int(x.b[0x070D]),
		int(x.b[0x0710]),
		int(x.b[0x0711]),
		int(x.b[0x0712]), 0, time.UTC)
}

func (x *Bytes) ProductType() string {
	const offset = 0x060B
	n := offset
	for ; n < offset+50; n++ {
		if x.b[n] == 0xff || x.b[n] == 0 {
			break
		}
	}
	return string(x.b[offset:n])
}

func (x *Bytes) Info() ProductFirmwareInfo {
	r := ProductFirmwareInfo{
		TempPoints:  x.TempPoints(),
		Time:        x.Time(),
		ProductType: x.ProductType(),
		Serial:      formatBCD(x.b[0x0701:0x0705]),
		Scale:       formatBCD(x.b[0x0602:0x0606]) + " - " + formatBCD(x.b[0x0606:0x060A]),
		Sensitivity: formatFloat(math.Float64frombits(binary.LittleEndian.Uint64(x.b[0x0720:]))),
	}
	for _, a := range x.units {
		if a.Code == x.b[0x060A] {
			r.Units = a.UnitsName
			break
		}
	}
	for _, a := range x.gases {
		if a.Code == x.b[0x0600] {
			r.Gas = a.GasName
			break
		}
	}
	return r
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func formatBCD(b []byte) string {
	if v, ok := goutils.ParseBCD6(b); ok {
		return formatFloat(v)
	} else {
		return fmt.Sprintf("% X", b)
	}
}

func (x *Bytes) TempPoints() (r TempPoints) {

	valAt := func(i int) float64 {
		a := binary.LittleEndian.Uint16(x.b[i:])
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
