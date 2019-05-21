package data

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm/modbus"
	"math"
	"strconv"
	"time"
)

type Firmware struct {
	Place     int
	CreatedAt time.Time
	Serial,
	ScaleBegin,
	ScaleEnd,
	KSens20 float64
	Fon, Sens   TableXY
	Gas, Units  byte
	ProductType string
}

type FirmwareInfo struct {
	TempPoints
	Place int
	Time  time.Time
	Sensitivity,
	Serial,
	ProductType,
	Gas,
	Units,
	ScaleBeg,
	ScaleEnd,
	ISMinus20,
	ISPlus20,
	ISPlus50 string
}

const FirmwareSize = 1832

type FirmwareBytes []byte

func (s Product) FirmwareBytes() (b FirmwareBytes, err error) {
	if len(s.Firmware) == 0 {
		err = merry.New("ЭХЯ не \"прошита\"")
		return
	}
	if len(s.Firmware) < FirmwareSize {
		err = merry.New("не верный формат \"прошивки\"")
		return
	}
	b = s.Firmware
	return
}

func (s ProductInfo) FirmwareInfo() FirmwareInfo {
	x := FirmwareInfo{
		Place:       s.Place,
		Gas:         s.GasName,
		Units:       s.UnitsName,
		ScaleBeg:    "0",
		ScaleEnd:    fmt.Sprintf("%v", s.Scale),
		ProductType: s.AppliedProductTypeName,
		Serial:      formatNullInt64(s.Serial),
		Time:        s.CreatedAt,
		Sensitivity: formatNullFloat64(s.KSens20, 3),
		ISPlus20:    formatNullFloat64K(s.ISPlus20, 1000, -1),
		ISMinus20:   formatNullFloat64K(s.ISMinus20, 1000, -1),
		ISPlus50:    formatNullFloat64K(s.ISPlus50, 1000, -1),
	}

	if fonM, err := s.TableFon(); err == nil {
		if sensM, err := s.TableSens(); err == nil {
			for k := range fonM {
				fonM[k] *= 1000
			}
			x.TempPoints = NewTempPoints(fonM, sensM)
		}
	}
	return x
}

func (s ProductInfo) Firmware() (x Firmware, err error) {

	if !s.Serial.Valid {
		err = merry.New("не задан серийный номер")
		return
	}
	if !s.KSens20.Valid {
		err = merry.New("нет значения к-та чувствительности")
		return
	}

	x = Firmware{
		Place:       s.Place,
		CreatedAt:   s.CreatedAt,
		ProductType: s.AppliedProductTypeName,
		Serial:      float64(s.Serial.Int64),
		KSens20:     s.KSens20.Float64,
		ScaleBegin:  0,
		ScaleEnd:    s.Scale,
		Gas:         s.GasCode,
		Units:       s.UnitsCode,
	}

	if x.Fon, err = s.TableFon(); err != nil {
		return
	}
	for k := range x.Fon {
		x.Fon[k] *= 1000
	}
	if x.Sens, err = s.TableSens(); err != nil {
		return
	}
	return
}

func (s ProductInfo) TableFon2() (y TableXY, err error) {
	y = TableXY{}

	y[20], err = s.CurrentValue(20, Fon)
	if err != nil {
		return
	}
	y[50], err = s.CurrentValue(50, Fon)
	if err != nil {
		return
	}

	y[40] = (y[50]-y[20])*0.5 + y[20]
	y[-40] = 0
	y[-20] = y[20] * 0.2
	y[0] = y[20] * 0.5
	y[30] = (y[40]-y[20])*0.5 + y[20]
	y[45] = (y[50]-y[40])*0.5 + y[40]

	return
}

func (s ProductInfo) TableFon3() (y TableXY, err error) {
	y = TableXY{}
	y[-20], err = s.CurrentValue(-20, Fon)
	if err != nil {
		return
	}
	y[20], err = s.CurrentValue(20, Fon)
	if err != nil {
		return
	}
	y[50], err = s.CurrentValue(50, Fon)
	if err != nil {
		return
	}
	y[40] = (y[50]-y[20])*0.5 + y[20]
	y[-40] = y[-20] - 0.5*(y[20]-y[-20])
	y[0] = y[20] - 0.5*(y[20]-y[-20])
	y[30] = (y[40]-y[20])*0.5 + y[20]
	y[45] = (y[50]-y[40])*0.5 + y[40]
	return
}

func (s ProductInfo) TableSens2() (TableXY, error) {
	y, err := s.KSensPercentValues(false)
	if err == nil {
		y[40] = (y[50]-y[20])*0.5 + y[20]
		y[-40] = 30
		y[-20] = 58
		y[0] = 82
		y[30] = (y[40]-y[20])*0.5 + y[20]
		y[45] = (y[50]-y[40])*0.5 + y[40]
	}
	return y, err
}

func (s ProductInfo) TableSens3() (TableXY, error) {
	y, err := s.KSensPercentValues(true)
	if err == nil {
		//if y[-20] > 0 && y[-20] < 0.45*y[20] {
		//	return y, errors.Errorf(
		//		"ток чувствительности: I(-20)=%v, I(+20)=%v, I(-20)>0, I(-20)<0.45*I(+20)",
		//		y[-20], y[20])
		//}
		y[0] = (y[20]-y[-20])*0.5 + y[-20]
		y[40] = (y[50]-y[20])*0.5 + y[20]
		y[45] = (y[50]-y[40])*0.5 + y[40]
		y[30] = (y[40]-y[20])*0.5 + y[20]
		y[-40] = 2*y[-20] - y[0]
		if y[-20] > 0 {
			y[-40] += 1.2 * (45 - y[-20]) / (0.43429 * math.Log(y[-20]))
		}
	}
	return y, err
}

func (s ProductInfo) TableSens() (TableXY, error) {
	switch s.AppliedPointsMethod {
	case 2:
		return s.TableSens2()
	case 3:
		return s.TableSens3()
	default:
		panic(fmt.Sprintf("wrong points method: %d", s.AppliedPointsMethod))
	}
}

func (s ProductInfo) TableFon() (TableXY, error) {
	switch s.AppliedPointsMethod {
	case 2:
		return s.TableFon2()
	case 3:
		return s.TableFon3()
	default:
		panic(fmt.Sprintf("wrong points method: %d", s.AppliedPointsMethod))
	}
}

func (x FirmwareBytes) Time() time.Time {
	_ = x[0x0712]
	return time.Date(
		2000+int(x[0x070F]),
		time.Month(x[0x070E]),
		int(x[0x070D]),
		int(x[0x0710]),
		int(x[0x0711]),
		int(x[0x0712]), 0, time.UTC)
}

func (x FirmwareBytes) ProductType() string {
	const offset = 0x060B
	n := offset
	for ; n < offset+50; n++ {
		if x[n] == 0xff || x[n] == 0 {
			break
		}
	}
	return string(x[offset:n])
}

func (x FirmwareBytes) FirmwareInfo(place int) FirmwareInfo {
	r := FirmwareInfo{
		Place:       place,
		TempPoints:  x.TempPoints(),
		Time:        x.Time(),
		ProductType: x.ProductType(),
		Serial:      formatBCD(x[0x0701:0x0705], -1),
		ScaleBeg:    formatBCD(x[0x0602:0x0606], -1),
		ScaleEnd:    formatBCD(x[0x0606:0x060A], -1),
		Sensitivity: formatFloat(math.Float64frombits(binary.LittleEndian.Uint64(x[0x0720:])), 3),
	}
	for _, a := range ListUnits() {
		if a.Code == x[0x060A] {
			r.Units = a.UnitsName
			break
		}
	}
	for _, a := range Gases() {
		if a.Code == x[0x0600] {
			r.Gas = a.GasName
			break
		}
	}
	return r
}

func (x FirmwareBytes) TempPoints() (r TempPoints) {

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

func (x Firmware) Bytes() (b FirmwareBytes) {

	b = make(FirmwareBytes, FirmwareSize)

	for i := 0; i < len(b); i++ {
		b[i] = 0xFF
	}

	modbus.PutBCD6(b[0x0701:], float64(x.Serial))
	modbus.PutBCD6(b[0x0602:], x.ScaleBegin)
	modbus.PutBCD6(b[0x0606:], x.ScaleEnd)

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
	binary.LittleEndian.PutUint64(b[0x0720:], math.Float64bits(x.KSens20))

	putTempValue := func(value float64, i int) {
		y := math.Round(value)
		n := uint16(y)
		binary.LittleEndian.PutUint16(b[i:], n)
	}

	at := NewApproximationTable(x.Fon)
	t := float64(-124)
	for i := 0x00F8; i >= 0; i -= 2 {
		putTempValue(at.F(t), i)
		t++
	}
	t = 0
	for i := 0x0100; i <= 0x01F8; i += 2 {
		putTempValue(at.F(t), i)
		t++
	}

	at = NewApproximationTable(x.Sens)
	t = float64(-124)
	for i := 0x04F8; i >= 0x0400; i -= 2 {
		putTempValue(at.F(t), i)
		t++
	}
	t = 0
	for i := 0x0500; i <= 0x05F8; i += 2 {
		putTempValue(at.F(t), i)
		t++
	}
	return
}

func formatNullInt64(v sql.NullInt64) string {
	if v.Valid {
		return strconv.FormatInt(v.Int64, 10)
	}
	return ""
}

func formatNullFloat64K(v sql.NullFloat64, k float64, precision int) string {
	if v.Valid {
		return formatFloat(v.Float64*k, precision)
	}
	return ""
}

func formatNullFloat64(v sql.NullFloat64, precision int) string {
	return formatNullFloat64K(v, 1, precision)
}

func formatFloat(v float64, precision int) string {
	return strconv.FormatFloat(v, 'f', precision, 64)
}

func formatBCD(b []byte, precision int) string {
	if v, ok := modbus.ParseBCD6(b); ok {
		return formatFloat(v, precision)
	} else {
		return fmt.Sprintf("% X", b)
	}
}
