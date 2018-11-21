package firmware

import (
	"encoding/binary"
	"fmt"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/goutils"
	"github.com/pkg/errors"
	"math"
	"time"
)

type Bytes []byte

type CodeStr struct {
	Code byte
	Str  string
}

func (x Bytes) PutProduct(p data.ProductInfo) error {
	if !p.Serial.Valid {
		return errors.New("не задан серийный номер")
	}
	if !p.KSens20.Valid {
		return errors.New("нет значения к-та чувствительности")
	}
	goutils.PutBCD6(x[0x0701:], float64(p.Serial.Int64))
	x[0x070F] = byte(p.CreatedAt.Year() - 2000)
	x[0x070E] = byte(p.CreatedAt.Month())
	x[0x070D] = byte(p.CreatedAt.Day())
	x[0x0710] = byte(p.CreatedAt.Hour())
	x[0x0711] = byte(p.CreatedAt.Minute())
	x[0x0712] = byte(p.CreatedAt.Second())
	x[0x0600] = p.GasCode
	x[0x060A] = p.UnitsCode
	goutils.PutBCD6(x[0x0602:], 0)
	goutils.PutBCD6(x[0x0606:], p.Scale)
	x.putProductTypeName(p.AppliedProductTypeName)
	binary.LittleEndian.PutUint64(x[0x0720:], math.Float64bits(p.KSens20.Float64))
	if err := x.putProductFonPoints(p); err != nil {
		return err
	}
	if err := x.putProductSensPoints(p); err != nil {
		return err
	}
	return nil
}

// putProductTypeName - исполнение
func (x Bytes) putProductTypeName(productTypeName string) {
	const n = 50
	b := []byte(productTypeName)
	if len(b) > n {
		b = b[:n]
	}
	for i := copy(x[0x060B:], b); i < n; i++ {
		x[i] = 0xFF
	}
}

// putFonPoints - записать в буфер точки фоновых токов
func (x Bytes) putProductFonPoints(p data.ProductInfo) error {
	xy, err := srcFon(p)
	if err != nil {
		return err
	}
	at := newApproxTbl(xy)
	t := float64(-124)
	for i := 0x00F8; i >= 0; i -= 2 {
		y := math.Round(at.F(t))
		n := uint16(y)
		binary.LittleEndian.PutUint16(x[i:], n)
		t++
	}
	return nil
}

// putSensPoints - записать в буфер точки коэфф. чувствительности
func (x Bytes) putProductSensPoints(p data.ProductInfo) error {
	xy, err := srcSens(p)
	if err != nil {
		return err
	}
	at := newApproxTbl(xy)
	t := float64(0)
	for i := 0x0100; i <= 0x01F8; i += 2 {
		y := math.Round(at.F(t))
		n := uint16(y)
		binary.LittleEndian.PutUint16(x[i:], n)
		t++
	}
	return nil
}

func (x Bytes) Date() time.Time {
	_ = x[0x0712]
	return time.Date(
		2000+int(x[0x070F]),
		time.Month(x[0x070E]),
		int(x[0x070D]),
		int(x[0x0710]),
		int(x[0x0711]),
		int(x[0x0712]), 0, time.UTC)
}

func (x Bytes) Sensitivity() float64 {
	bits := binary.LittleEndian.Uint64(x[0x0720:])
	return math.Float64frombits(bits)
}

func (x Bytes) Serial() (serial float64) {
	serial, _ = goutils.ParseBCD6(x[0x0701:])
	return
}

func (x Bytes) ProductType() string {
	const offset = 0x060B
	n := offset
	for ; n < offset+50; n++ {
		if x[offset+n] == 0xff || x[offset+n] == 0 {
			break
		}
	}
	return string(x[offset:n])
}

func (x Bytes) String() string {

	serial, ok := goutils.ParseBCD6(x[0x0701:])
	if !ok {
		return "?"
	}

	var bs []byte
	for i := 0x060B; i < 0x060B+50; i++ {
		if x[i] == 0xff {
			break
		}
		bs = append(bs, x[i])
	}
	{
		n := len(bs)
		if n > 0 && bs[n-1] == 0 {
			bs = bs[:n-1]
		}
	}

	return fmt.Sprintf("№%v %s %.3f", serial, string(bs), x.Sensitivity())
}

func (x Bytes) FonPoints() (xs, ys []float64) {
	t := float64(-124)
	for i := 0x00F8; i >= 0; i -= 2 {
		if x[i] == 0xFF && x[i+1] == 0xFF {
			continue
		}
		xs = append(xs, t)
		a := binary.LittleEndian.Uint16(x[i:])
		b := int16(a)
		y := float64(b)
		ys = append(ys, y)
		t++
	}
	t = 0
	for i := 0x0100; i <= 0x01F8; i += 2 {
		if x[i] == 0xFF && x[i+1] == 0xFF {
			continue
		}
		xs = append(xs, t)
		a := binary.LittleEndian.Uint16(x[i:])
		b := int16(a)
		y := float64(b)
		ys = append(ys, y)
		t++
	}
	return
}

func (x Bytes) SensPoints() (xs, ys []float64) {
	t := float64(-124)
	for i := 0x04F8; i >= 0x0400; i -= 2 {
		if x[i] == 0xFF && x[i+1] == 0xFF {
			continue
		}
		xs = append(xs, t)
		a := binary.LittleEndian.Uint16(x[i:])
		b := int16(a)
		y := float64(b)
		ys = append(ys, y)
		t++
	}
	t = 0
	for i := 0x0500; i <= 0x05F8; i += 2 {
		if x[i] == 0xFF && x[i+1] == 0xFF {
			continue
		}
		xs = append(xs, t)
		a := binary.LittleEndian.Uint16(x[i:])
		b := int16(a)
		y := float64(b)
		ys = append(ys, y)
		t++
	}
	return
}
