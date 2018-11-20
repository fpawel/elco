package firmware

import (
	"encoding/binary"
	"fmt"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/goutils"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"math"
	"time"
)

type Bytes []byte

type CodeStr struct {
	Code byte
	Str  string
}

func (x Bytes) Put(p data.ProductInfo1) (err error) {

	// дата создания
	x[0x070F] = byte(p.CreatedAt.Year() - 2000)
	x[0x070E] = byte(p.CreatedAt.Month())
	x[0x070D] = byte(p.CreatedAt.Day())
	x[0x0710] = byte(p.CreatedAt.Hour())
	x[0x0711] = byte(p.CreatedAt.Minute())
	x[0x0712] = byte(p.CreatedAt.Second())

	// серийный номер
	if p.Serial.Valid {
		goutils.PutBCD6(x[0x0701:], float64(p.Serial.Int64))
	} else {
		err = multierror.Append(err, errors.New("не задан серийный номер"))
	}
	// коэффициент чувствительность
	if p.KSens20.Valid {
		binary.LittleEndian.PutUint64(x[0x0720:], math.Float64bits(p.KSens20.Float64))
	} else {
		err = multierror.Append(err, errors.New("нет значения Кч"))
	}
	x[0x0600] = p.GasCode                // газ
	x[0x060A] = p.UnitsCode              // единицы измерения
	goutils.PutBCD6(x[0x0602:], 0)       // начало шкалы
	goutils.PutBCD6(x[0x0606:], p.Scale) // конец шкалы
	{                                    // исполнение
		const n = 50
		b := []byte(p.ProductTypeName)
		if len(b) > n {
			b = b[:n]
		}
		for i := copy(x[0x060B:], b); i < n; i++ {
			x[i] = 0xFF
		}
	}

	// фоновые токи
	t := float64(-124)
	for i := 0x00F8; i >= 0; i -= 2 {
		binary.LittleEndian.PutUint16(x[i:], uint16(t))

		xs = append(xs, t)
		a := binary.LittleEndian.Uint16(x[i:])
		b := int16(a)
		y := float64(b)
		ys = append(ys, y)
		t++
	}
	t = 0
	for i := 0x0100; i <= 0x01F8; i += 2 {
		xs = append(xs, t)
		a := binary.LittleEndian.Uint16(x[i:])
		b := int16(a)
		y := float64(b)
		ys = append(ys, y)
		t++
	}

	return
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

func (x Bytes) CurrFonPoints() (xs, ys []float64) {
	if len(x) < 0x01F8+1 {
		return
	}
	var t float64 = -124
	for i := 0x00F8; i >= 0; i -= 2 {
		xs = append(xs, t)
		a := binary.LittleEndian.Uint16(x[i:])
		b := int16(a)
		y := float64(b)
		ys = append(ys, y)
		t++
	}
	t = 0
	for i := 0x0100; i <= 0x01F8; i += 2 {
		xs = append(xs, t)
		a := binary.LittleEndian.Uint16(x[i:])
		b := int16(a)
		y := float64(b)
		ys = append(ys, y)
		t++
	}
	return
}

func (x Bytes) CurrSensPoints() (xs, ys []float64) {

	if len(x) < 0x05F8+1 {
		return
	}

	var t float64 = -124
	for i := 0x04F8; i >= 0x0400; i -= 2 {
		xs = append(xs, t)
		a := binary.LittleEndian.Uint16(x[i:])
		b := int16(a)
		y := float64(b)
		ys = append(ys, y)
		t++
	}
	t = 0
	for i := 0x0500; i <= 0x05F8; i += 2 {
		xs = append(xs, t)
		a := binary.LittleEndian.Uint16(x[i:])
		b := int16(a)
		y := float64(b)
		ys = append(ys, y)
		t++
	}
	return
}
