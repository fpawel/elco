package api

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/crud"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/goutils"
	"github.com/pkg/errors"
	"regexp"
	"strings"
	"time"
)

type ProductFirmware struct {
	c crud.ProductFirmware
	w WriteSingleProductFirmwareRunner
}

type WriteSingleProductFirmwareRunner interface {
	RunWriteSingleProductFirmware(number int, bytes []byte)
}

type FirmwareInfo2 struct {
	Number                                 int
	Year, Month, Day, Hour, Minute, Second int
	Sensitivity,
	Serial,
	ProductType,
	Gas,
	Units,
	Scale string
	Values []string
}

type TempValues struct {
	Values []string
}

func NewProductFirmware(c crud.ProductFirmware, w WriteSingleProductFirmwareRunner) *ProductFirmware {
	return &ProductFirmware{c, w}
}

func (x *ProductFirmware) StoredFirmwareInfo(productID [1]int64, r *data.FirmwareInfo) error {
	if b, err := x.c.StoredFirmwareInfo(productID[0]); err != nil {
		return err
	} else {
		*r = b.FirmwareInfo(x.c.ListUnits(), x.c.ListGases())
		return nil
	}
}

func (x *ProductFirmware) CalculateFirmwareInfo(productID [1]int64, r *data.FirmwareInfo) (err error) {
	*r, err = x.c.CalculateFirmwareInfo(productID[0])
	return
}

func (x *ProductFirmware) TempPoints(v TempValues, r *data.TempPoints) error {
	fonM, sensM := data.TableXY{}, data.TableXY{}
	if err := tempPoints(v.Values, fonM, sensM); err != nil {
		return err
	}
	*r = data.NewTempPoints(fonM, sensM)
	return nil
}

func (x *ProductFirmware) Write(v FirmwareInfo2, _ *struct{}) (err error) {

	z := data.Firmware{
		CreatedAt: time.Date(v.Year, time.Month(v.Month), v.Day, v.Hour, v.Minute, v.Second, 0, time.Local),
	}

	ok := false
	gases := x.c.ListGases()
	for _, gas := range gases {
		if gas.GasName == v.Gas {
			z.Gas = gas.Code
			ok = true
			break
		}
	}
	if !ok {
		return merry.Errorf("код газа не задан: %q ", v.Gas)
	}

	ok = false
	units := x.c.ListUnits()
	for _, u := range units {
		if u.UnitsName == v.Units {
			z.Units = u.Code
			ok = true
			break
		}
	}
	if !ok {
		return merry.Errorf("код единиц измерения не задан: %q ", v.Units)
	}

	m := regexp.MustCompile(`\s*(\d+)\s*-\s*(\d+)\s*`).FindStringSubmatch(v.Scale)
	if len(m) < 3 {
		return merry.Appendf(err, "не верный формат значения шкалы: %s", v.Scale)
	}

	z.Scale, err = goutils.ParseFloat(m[2])
	if err != nil {
		return merry.Appendf(err, "не верный формат значения шкалы: %s", m[2])
	}

	z.Fon, z.Sens = data.TableXY{}, data.TableXY{}

	if err = tempPoints(v.Values, z.Fon, z.Sens); err != nil {
		return merry.Append(err, "не удался расчёт температурных точек")
	}

	z.KSens20 = data.NewApproximationTable(z.Sens).F(20)

	b := z.Bytes()
	x.w.RunWriteSingleProductFirmware(v.Number, b[:])

	return
}

func tempPoints(values []string, fonM data.TableXY, sensM data.TableXY) error {
	if len(values)%3 != 0 {
		return errors.New("sequence length is not a multiple of three")
	}

	for n := 0; n < len(values); n += 3 {
		strT := strings.TrimSpace(values[n+0])
		if len(strT) == 0 {
			continue
		}

		t, err := goutils.ParseFloat(values[n])
		if err != nil {
			return merry.Appendf(err, "строка %d", n)
		}
		strI := strings.TrimSpace(values[n+1])
		if len(strI) > 0 {
			var i float64
			i, err = goutils.ParseFloat(strI)
			if err != nil {
				return merry.Appendf(err, "строка %d", n)
			}
			fonM[t] = i
		}
		strS := strings.TrimSpace(values[n+2])
		if len(strS) > 0 {
			var k float64
			k, err = goutils.ParseFloat(strS)
			if err != nil {
				return merry.Appendf(err, "строка %d", n)
			}
			sensM[t] = k
		}
	}
	// r = data.NewTempPoints(fonM, sensM)
	return nil
}
