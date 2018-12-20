package firmware

import (
	"encoding/binary"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/goutils"
	"math"
	"time"
)

const Size = 1832

type Bytes struct {
	gases []data.Gas
	units []data.Units
	B     [Size]byte
}

type CodeStr struct {
	Code byte
	Str  string
}

func FromBytes(b []byte, gases []data.Gas, units []data.Units) (*Bytes, error) {
	if len(b) == 0 {
		return nil, merry.New("ЭХЯ не \"прошита\"")
	}
	if len(b) < Size {
		return nil, merry.New("не верный формат \"прошивки\"")
	}
	x := &Bytes{
		gases: gases,
		units: units,
	}
	copy(x.B[:Size], b)
	return x, nil
}

func FromProductInfo(p data.ProductInfo, gases []data.Gas, units []data.Units) (*Bytes, error) {
	x := &Bytes{
		gases: gases,
		units: units,
	}
	if !p.Serial.Valid {
		return x, merry.New("не задан серийный номер")
	}
	if !p.KSens20.Valid {
		return x, merry.New("нет значения к-та чувствительности")
	}
	for i := 0; i < len(x.B); i++ {
		x.B[i] = 0xFF
	}
	goutils.PutBCD6(x.B[0x0701:], float64(p.Serial.Int64))
	goutils.PutBCD6(x.B[0x0602:], 0)
	goutils.PutBCD6(x.B[0x0606:], p.Scale)

	x.B[0x070F] = byte(p.CreatedAt.Year() - 2000)
	x.B[0x070E] = byte(p.CreatedAt.Month())
	x.B[0x070D] = byte(p.CreatedAt.Day())
	x.B[0x0710] = byte(p.CreatedAt.Hour())
	x.B[0x0711] = byte(p.CreatedAt.Minute())
	x.B[0x0712] = byte(p.CreatedAt.Second())
	x.B[0x0600] = p.GasCode
	x.B[0x060A] = p.UnitsCode
	x.PutProductTypeName(p.AppliedProductTypeName)
	binary.LittleEndian.PutUint64(x.B[0x0720:], math.Float64bits(p.KSens20.Float64))
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
	for i := copy(x.B[0x060B:], b); i < n; i++ {
		x.B[i] = 0xFF
	}
}

func (x *Bytes) putTempValue(value float64, i int) {
	y := math.Round(value)
	n := uint16(y)
	binary.LittleEndian.PutUint16(x.B[i:], n)
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
	_ = x.B[0x0712]
	return time.Date(
		2000+int(x.B[0x070F]),
		time.Month(x.B[0x070E]),
		int(x.B[0x070D]),
		int(x.B[0x0710]),
		int(x.B[0x0711]),
		int(x.B[0x0712]), 0, time.UTC)
}

func (x *Bytes) ProductType() string {
	const offset = 0x060B
	n := offset
	for ; n < offset+50; n++ {
		if x.B[n] == 0xff || x.B[n] == 0 {
			break
		}
	}
	return string(x.B[offset:n])
}

func (x *Bytes) Info() ProductFirmwareInfo {
	r := ProductFirmwareInfo{
		TempPoints:  x.TempPoints(),
		Time:        x.Time(),
		ProductType: x.ProductType(),
		Serial:      formatBCD(x.B[0x0701:0x0705], -1),
		Scale:       formatBCD(x.B[0x0602:0x0606], -1) + " - " + formatBCD(x.B[0x0606:0x060A], -1),
		Sensitivity: formatFloat(math.Float64frombits(binary.LittleEndian.Uint64(x.B[0x0720:])), 3),
	}
	for _, a := range x.units {
		if a.Code == x.B[0x060A] {
			r.Units = a.UnitsName
			break
		}
	}
	for _, a := range x.gases {
		if a.Code == x.B[0x0600] {
			r.Gas = a.GasName
			break
		}
	}
	return r
}

func (x *Bytes) TempPoints() (r TempPoints) {

	valAt := func(i int) float64 {
		a := binary.LittleEndian.Uint16(x.B[i:])
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
