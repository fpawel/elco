package api

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/data"
	"github.com/pkg/errors"
	"gopkg.in/reform.v1"
	"strconv"
	"strings"
	"time"
)

type ProductFirmware struct {
	db *reform.DB
	f  FirmwareRunner
}

type FirmwareRunner interface {
	RunWriteFirmware(place int, bytes []byte)
	RunReadFirmware(place int)
}

type FirmwareInfo2 struct {
	Place,
	Year, Month, Day, Hour, Minute, Second int
	Sensitivity, Serial, ProductType,
	Gas, Units, ScaleBegin, ScaleEnd string
	Values []string
}

type TempValues struct {
	Values []string
}

func NewProductFirmware(db *reform.DB, f FirmwareRunner) *ProductFirmware {
	return &ProductFirmware{db, f}
}

func (x *ProductFirmware) StoredFirmwareInfo(productID [1]int64, r *data.FirmwareInfo) error {

	var p data.Product
	if err := x.db.SelectOneTo(&p, `WHERE product_id = ?`, productID[0]); err != nil {
		return err
	}
	if len(p.Firmware) == 0 {
		return merry.New("ЭХЯ не \"прошита\"")
	}
	if len(p.Firmware) < data.FirmwareSize {
		return merry.New("не верный формат \"прошивки\"")
	}
	gases, err := data.ListGases(x.db)
	if err != nil {
		return err
	}
	units, err := data.ListUnits(x.db)
	if err != nil {
		return err
	}
	*r = data.FirmwareBytes(p.Firmware).FirmwareInfo(units, gases)
	r.Place = p.Place
	return nil
}

func (x *ProductFirmware) CalculateFirmwareInfo(productID [1]int64, r *data.FirmwareInfo) (err error) {
	var p data.ProductInfo
	if err := x.db.SelectOneTo(&p, `WHERE product_id = ?`, productID[0]); err != nil {
		return err
	}
	*r = p.FirmwareInfo()
	return nil
}

func (x *ProductFirmware) TempPoints(v TempValues, r *data.TempPoints) error {
	fonM, sensM := data.TableXY{}, data.TableXY{}
	if err := tempPoints(v.Values, fonM, sensM); err != nil {
		return err
	}
	*r = data.NewTempPoints(fonM, sensM)
	return nil
}

func (x *ProductFirmware) RunReadFirmware(place [1]int, _ *struct{}) error {
	x.f.RunReadFirmware(place[0])
	return nil
}

func (x *ProductFirmware) RunWriteFirmware(v FirmwareInfo2, _ *struct{}) (err error) {

	z := data.Firmware{
		CreatedAt: time.Date(v.Year, time.Month(v.Month), v.Day, v.Hour, v.Minute, v.Second, 0, time.Local),
	}

	gases, err := data.ListGases(x.db)
	if err != nil {
		return err
	}
	units, err := data.ListUnits(x.db)
	if err != nil {
		return err
	}

	ok := false

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

	z.ScaleBegin, err = parseFloat(v.ScaleBegin)
	if err != nil {
		return merry.Appendf(err, "не верный формат значения начала шкалы: %s", v.ScaleBegin)
	}

	z.ScaleEnd, err = parseFloat(v.ScaleEnd)
	if err != nil {
		return merry.Appendf(err, "не верный формат значения конца шкалы: %s", v.ScaleEnd)
	}

	z.Fon, z.Sens = data.TableXY{}, data.TableXY{}

	if err = tempPoints(v.Values, z.Fon, z.Sens); err != nil {
		return merry.Append(err, "не удался расчёт температурных точек")
	}

	z.KSens20 = data.NewApproximationTable(z.Sens).F(20)

	x.f.RunWriteFirmware(v.Place, z.Bytes())
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

		t, err := parseFloat(values[n])
		if err != nil {
			return merry.Appendf(err, "строка %d", n)
		}
		strI := strings.TrimSpace(values[n+1])
		if len(strI) > 0 {
			var i float64
			i, err = parseFloat(strI)
			if err != nil {
				return merry.Appendf(err, "строка %d", n)
			}
			fonM[t] = i
		}
		strS := strings.TrimSpace(values[n+2])
		if len(strS) > 0 {
			var k float64
			k, err = parseFloat(strS)
			if err != nil {
				return merry.Appendf(err, "строка %d", n)
			}
			sensM[t] = k
		}
	}
	// r = data.NewTempPoints(fonM, sensM)
	return nil
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.Replace(s, ",", ".", -1), 64)
}
