package chipmem

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/data"
	"time"
)

type FirmwareInfo struct {
	ProductTempPoints                      TempPoints
	TempValues                             []string
	Year, Month, Day, Hour, Minute, Second int
	SensitivityLab73,
	SensitivityProduct,
	Fon20,
	Serial,
	ProductType,
	Gas,
	Units,
	ScaleBeg,
	ScaleEnd string
}

func (v FirmwareInfo) GetFirmware() (Firmware, error) {
	z := Firmware{
		CreatedAt: time.Date(v.Year, time.Month(v.Month), v.Day, v.Hour, v.Minute, v.Second, 0, time.Local),
	}

	gases := data.ListGases()
	units := data.ListUnits()

	ok := false

	for _, gas := range gases {
		if gas.GasName == v.Gas {
			z.Gas = gas.Code
			ok = true
			break
		}
	}
	if !ok {
		return z, merry.Errorf("код газа не задан: %q ", v.Gas)
	}

	ok = false

	for _, u := range units {
		if u.UnitsName == v.Units {
			z.Units = u.Code
			ok = true
			break
		}
	}
	if !ok {
		return z, merry.Errorf("код единиц измерения не задан: %q ", v.Units)
	}

	var err error
	z.ScaleBegin, err = parseFloat(v.ScaleBeg)
	if err != nil {
		return z, merry.Appendf(err, "не верный формат значения начала шкалы: %s", v.ScaleBeg)
	}

	z.ScaleEnd, err = parseFloat(v.ScaleEnd)
	if err != nil {
		return z, merry.Appendf(err, "не верный формат значения конца шкалы: %s", v.ScaleEnd)
	}

	z.Fon, z.Sens = data.TableXY{}, data.TableXY{}

	if err = GetTempTables(v.TempValues, z.Fon, z.Sens); err != nil {
		return z, merry.Append(err, "не удался расчёт температурных точек")
	}

	z.KSens20, err = parseFloat(v.SensitivityProduct)
	if err != nil {
		return z, merry.Appendf(err, "не верный формат значения коэффициента чувствиттельности: %s", v.SensitivityProduct)
	}

	z.Fon20, err = parseFloat(v.Fon20)
	if err != nil {
		return z, merry.Appendf(err, "не верный формат значения фонового тока: %s", v.Fon20)
	}

	z.ProductType = v.ProductType

	z.Serial, err = parseFloat(v.Serial)
	if err != nil {
		return z, merry.Appendf(err, "не верный формат значения серийного номера: %s", v.Serial)
	}

	return z, nil
}

func (v FirmwareInfo) CalculateBytes() ([]string, error) {
	firmware, err := v.GetFirmware()
	if err != nil {
		return nil, err
	}
	var xs []string
	for _, b := range firmware.Bytes() {
		xs = append(xs, fmt.Sprintf("%02X", b))
	}
	return xs, nil
}
