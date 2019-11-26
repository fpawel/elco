package chipmem

import (
	"encoding/binary"
	"github.com/fpawel/comm/modbus"
	"math"
	"time"
)

const FirmwareSize = 1832

type Firmware struct {
	Place     int
	CreatedAt time.Time
	Serial,
	ScaleBegin,
	ScaleEnd,
	Fon20,
	KSens20 float64
	Fon, Sens   TableXY
	Gas, Units  byte
	ProductType string
}

func (x Firmware) Bytes() (b Bytes) {

	b = make(Bytes, FirmwareSize)

	for i := 0; i < len(b); i++ {
		b[i] = 0xFF
	}

	modbus.PutBCD6(b[0x0701:0x0705], x.Serial)
	modbus.PutBCD6(b[0x0602:0x0606], x.ScaleBegin)
	modbus.PutBCD6(b[0x0606:0x060A], x.ScaleEnd)

	b[0x070F] = byte(x.CreatedAt.Year() - 2000)
	b[0x070E] = byte(x.CreatedAt.Month())
	b[0x070D] = byte(x.CreatedAt.Day())
	b[0x0710] = byte(x.CreatedAt.Hour())
	b[0x0711] = byte(x.CreatedAt.Minute())
	b[0x0712] = byte(x.CreatedAt.Second())
	b[0x0600] = x.Gas
	b[0x060A] = x.Units

	bProductType := []byte(x.ProductType)
	if len(bProductType) > 50 {
		bProductType = bProductType[:50]
	}
	copy(b[0x060B:], bProductType)

	atFon := NewApproximationTable(x.Fon)

	binary.LittleEndian.PutUint64(b[0x0720:], math.Float64bits(x.KSens20))
	modbus.PutBCD6(b[0x0709:0x70D], x.KSens20)
	modbus.PutBCD6(b[0x0705:0x0709], x.Fon20)

	putTempValue := func(value float64, i int) {
		y := math.Round(value)
		n := uint16(y)
		binary.LittleEndian.PutUint16(b[i:], n)
	}

	t := float64(-124)
	for i := 0x00F8; i >= 0; i -= 2 {
		putTempValue(atFon.F(t), i)
		t++
	}
	t = 0
	for i := 0x0100; i <= 0x01F8; i += 2 {
		putTempValue(atFon.F(t), i)
		t++
	}

	atSens := NewApproximationTable(x.Sens)
	t = float64(-124)
	for i := 0x04F8; i >= 0x0400; i -= 2 {
		putTempValue(atSens.F(t), i)
		t++
	}
	t = 0
	for i := 0x0500; i <= 0x05F8; i += 2 {
		putTempValue(atSens.F(t), i)
		t++
	}
	return
}
