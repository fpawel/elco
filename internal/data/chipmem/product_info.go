package chipmem

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/data"
	"math"
)

type ProductInfo struct {
	P data.ProductInfo
}

func (s ProductInfo) CalculateFirmware() (x Firmware, err error) {

	if !s.P.Serial.Valid {
		err = merry.New("не задан серийный номер")
		return
	}
	if !s.P.KSens20.Valid {
		err = merry.New("нет значения к-та чувствительности")
		return
	}

	x = Firmware{
		Place:       s.P.Place,
		CreatedAt:   s.P.CreatedAt,
		ProductType: s.P.AppliedProductTypeName,
		Serial:      float64(s.P.Serial.Int64),
		KSens20:     s.P.KSens20.Float64,
		Fon20:       s.P.IFPlus20.Float64,
		ScaleBegin:  0,
		ScaleEnd:    s.P.Scale,
		Gas:         s.P.GasCode,
		Units:       s.P.UnitsCode,
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

	y[20], err = s.P.CurrentValue(20, data.Fon)
	if err != nil {
		return
	}
	y[50], err = s.P.CurrentValue(50, data.Fon)
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
	y[-20], err = s.P.CurrentValue(-20, data.Fon)
	if err != nil {
		return
	}
	y[20], err = s.P.CurrentValue(20, data.Fon)
	if err != nil {
		return
	}
	y[50], err = s.P.CurrentValue(50, data.Fon)
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
	y, err := s.P.KSensPercentValues(false)
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

	y, err := s.P.KSensPercentValues(true)
	if err == nil {

		if y[-20] <= 1 {
			return nil, merry.Errorf("y[-20]=%v: значение должно быть больше 1", y[-20])
		}

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
	switch s.P.AppliedPointsMethod {
	case 2:
		return s.TableSens2()
	case 3:
		return s.TableSens3()
	default:
		panic(fmt.Sprintf("wrong points method: %d", s.P.AppliedPointsMethod))
	}
}

func (s ProductInfo) TableFon() (TableXY, error) {
	switch s.P.AppliedPointsMethod {
	case 2:
		return s.TableFon2()
	case 3:
		return s.TableFon3()
	default:
		panic(fmt.Sprintf("wrong points method: %d", s.P.AppliedPointsMethod))
	}
}

func (s ProductInfo) FirmwareInfo() FirmwareInfo {
	t := s.P.CreatedAt
	x := FirmwareInfo{
		ProductTempPoints:  TempPoints{},
		TempValues:         []string{},
		Place:              s.P.Place,
		Year:               t.Year(),
		Month:              int(t.Month()),
		Day:                t.Day(),
		Hour:               t.Hour(),
		Minute:             t.Minute(),
		Second:             t.Second(),
		SensitivityLab73:   formatNullFloat64(s.P.KSens20, 3),
		SensitivityProduct: formatNullFloat64(s.P.KSens20, 3),
		Fon20:              formatNullFloat64K(s.P.IFPlus20, 1000, -1),
		Serial:             formatNullInt64(s.P.Serial),
		ProductType:        s.P.AppliedProductTypeName,
		Gas:                s.P.GasName,
		Units:              s.P.UnitsName,
		ScaleBeg:           "0",
		ScaleEnd:           fmt.Sprintf("%v", s.P.Scale),
	}

	if fonM, err := s.TableFon(); err == nil {
		if sensM, err := s.TableSens(); err == nil {
			for k := range fonM {
				fonM[k] *= 1000
			}
			x.ProductTempPoints = NewTempPoints(fonM, sensM)
			x.TempValues = tempValues(x.ProductTempPoints)
		}
	}
	return x
}
