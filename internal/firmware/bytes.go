package firmware

import (
	"encoding/binary"
	"fmt"
	"github.com/fpawel/goutils/serial/modbus"
	"math"
	"time"
)

type Bytes []byte

func (x Bytes) SetDate(t time.Time) {
	_ = x[0x0712]
	x[0x070F] = byte(t.Year() - 2000)
	x[0x070E] = byte(t.Month())
	x[0x070D] = byte(t.Day())
	x[0x0710] = byte(t.Hour())
	x[0x0711] = byte(t.Minute())
	x[0x0712] = byte(t.Second())
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
	if len(x) < 0x0720+1 {
		return float64(0)
	}
	return float64FromBytes(x[0x0720:])
}

func (x Bytes) Serial() (serial float64) {
	if len(x) < 0x0701+1 {
		return
	}
	serial, _ = modbus.ParseBCD(x[0x0701:])
	return
}

func (x Bytes) ProductType() string {
	if len(x) < 0x060B+50 {
		return ""
	}

	var bs []byte
	for i := 0x060B; i < 0x060B+50; i++ {
		if x[i] == 0xff {
			break
		}
		bs = append(bs, x[i])
	}

	if n := len(bs); n > 0 && bs[n-1] == 0 {
		bs = bs[:n-1]
	}
	return string(bs)
}

func (x Bytes) String() string {

	serial, ok := modbus.ParseBCD(x[0x0701:])
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

	return fmt.Sprintf("№%v %s %.3f", serial, string(bs), float64FromBytes(x[0x0720:]))
}

func float64FromBytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

func (x Bytes) SeriesFon() (xs, ys []float64) {
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

func (x Bytes) SeriesSens() (xs, ys []float64) {

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
