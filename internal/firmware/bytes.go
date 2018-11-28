package firmware

import (
	"encoding/binary"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/goutils"
	"github.com/pkg/errors"
	"math"
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

type TempPoints = struct {
	Temp, Fon, Sens [250]float64
}

type FlashInfo struct {
	TempPoints
	Time        time.Time
	Sensitivity float64
	Serial      float64
	ProductType,
	Gas,
	Units string
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

	if !p.Serial.Valid {
		return nil, errors.New("не задан серийный номер")
	}
	if !p.KSens20.Valid {
		return nil, errors.New("нет значения к-та чувствительности")
	}

	x := &Bytes{
		gases: gases,
		units: units,
	}
	for i := 0; i < len(x.b); i++ {
		x.b[i] = 0xFF
	}

	goutils.PutBCD6(x.b[0x0701:], float64(p.Serial.Int64))
	x.b[0x070F] = byte(p.CreatedAt.Year() - 2000)
	x.b[0x070E] = byte(p.CreatedAt.Month())
	x.b[0x070D] = byte(p.CreatedAt.Day())
	x.b[0x0710] = byte(p.CreatedAt.Hour())
	x.b[0x0711] = byte(p.CreatedAt.Minute())
	x.b[0x0712] = byte(p.CreatedAt.Second())
	x.b[0x0600] = p.GasCode
	x.b[0x060A] = p.UnitsCode
	goutils.PutBCD6(x.b[0x0602:], 0)
	goutils.PutBCD6(x.b[0x0606:], p.Scale)
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

//PutProductFonPoints - записать в буфер точки фоновых токов
func (x *Bytes) PutProductFonPoints(p data.ProductInfo) error {
	xy, err := srcFon(p)
	if err != nil {
		return err
	}
	at := newApproxTbl(xy)
	t := float64(-124)
	for i := 0x00F8; i >= 0; i -= 2 {
		y := math.Round(at.F(t))
		n := uint16(y)
		binary.LittleEndian.PutUint16(x.b[i:], n)
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
	t := float64(0)
	for i := 0x0100; i <= 0x01F8; i += 2 {
		y := math.Round(at.F(t))
		n := uint16(y)
		binary.LittleEndian.PutUint16(x.b[i:], n)
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

func (x *Bytes) Sensitivity() float64 {
	bits := binary.LittleEndian.Uint64(x.b[0x0720:])
	return math.Float64frombits(bits)
}

func (x *Bytes) Serial() float64 {

	if serial, ok := goutils.ParseBCD6(x.b[0x0701:]); ok {
		return serial
	}
	return math.NaN()
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

func (x *Bytes) Info() FlashInfo {
	r := FlashInfo{
		Serial:      x.Serial(),
		Sensitivity: x.Sensitivity(),
		Time:        x.Time(),
		ProductType: x.ProductType(),
		TempPoints:  x.TempPoints(),
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

func (x *Bytes) TempPoints() (r TempPoints) {
	t := float64(-124)
	n := 0
	for i := 0x00F8; i >= 0; i -= 2 {
		r.Temp[n] = t
		r.Fon[n] = x.tempValueAtAddr(i)
		t++
		n++
	}
	t = 0
	for i := 0x0100; i <= 0x01F8; i += 2 {
		r.Temp[n] = t
		r.Fon[n] = x.tempValueAtAddr(i)
		t++
		n++
	}
	t = -124
	n = 0
	for i := 0x04F8; i >= 0x0400; i -= 2 {
		r.Sens[n] = x.tempValueAtAddr(i)
		t++
		n++
	}
	t = 0
	for i := 0x0500; i <= 0x05F8; i += 2 {
		r.Sens[n] = x.tempValueAtAddr(i)
		t++
		n++
	}
	return
}

func (x *Bytes) tempValueAtAddr(i int) (v float64) {
	if x.b[i] == 0xFF && x.b[i+1] == 0xFF {
		return math.NaN()
	} else {
		a := binary.LittleEndian.Uint16(x.b[i:])
		b := int16(a)
		y := float64(b)
		return y
	}
}
