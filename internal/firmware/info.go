package firmware

import (
	"database/sql"
	"fmt"
	"github.com/fpawel/elco/internal/data"
	"strconv"
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

type TempPoints = struct {
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
		Sensitivity: formatNullFloat64(p.KSens20),
	}

	if fonM, err := srcFon(p); err == nil {
		if sensM, err := srcSens(p); err == nil {
			for k := range fonM {
				fonM[k] *= 1000
			}
			atFon := newApproxTbl(fonM)
			atSens := newApproxTbl(sensM)
			i := 0
			for t := float64(-124); t <= 125; t++ {
				x.Temp[i] = t
				x.Fon[i] = atFon.F(t)
				x.Sens[i] = atSens.F(t)
				i++
			}
		}
	}
	return x
}

func formatNullInt64(v sql.NullInt64) string {
	if v.Valid {
		return strconv.FormatInt(v.Int64, 10)
	}
	return ""
}

func formatNullFloat64(v sql.NullFloat64) string {
	if v.Valid {
		return formatFloat(v.Float64)
	}
	return ""
}
