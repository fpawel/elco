package firmware

import (
	"fmt"
	"github.com/fpawel/elco/internal/data"
	"time"
)

type ProductFirmwareInfo struct {
	TempPoints
	Time time.Time
	Sensitivity,
	Serial,
	ProductType,
	Gas,
	Units,
	Scale string
}

type TempPoints struct {
	Temp, Fon, Sens [250]float64
}

func CalculateProductFirmwareInfo(p data.ProductInfo) ProductFirmwareInfo {
	x := ProductFirmwareInfo{
		Gas:         p.GasName,
		Units:       p.UnitsName,
		Scale:       fmt.Sprintf("0 - %v", p.Scale),
		ProductType: p.AppliedProductTypeName,
		Serial:      formatNullInt64(p.Serial),
		Time:        p.CreatedAt,
		Sensitivity: formatNullFloat64(p.KSens20, 3),
	}

	if fonM, err := srcFon(p); err == nil {
		if sensM, err := srcSens(p); err == nil {
			for k := range fonM {
				fonM[k] *= 1000
			}
			x.TempPoints = CalculateTempPoints(fonM, sensM)
		}
	}
	return x
}

func minusOne(_ float64) float64 {
	return -1
}

func CalculateTempPoints(fonM, sensM M) (r TempPoints) {

	fFon := minusOne
	fSens := minusOne

	if len(fonM) > 0 {
		fFon = newApproxTbl(fonM).F
	}
	if len(sensM) > 0 {
		fSens = newApproxTbl(sensM).F
	}
	i := 0
	for t := float64(-124); t <= 125; t++ {
		r.Temp[i] = t
		r.Fon[i] = fFon(t)
		r.Sens[i] = fSens(t)
		i++
	}
	return
}
