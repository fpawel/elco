package api

import (
	"database/sql"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/data"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

type PlaceFirmware struct {
	f FirmwareRunner
}

type FirmwareRunner interface {
	RunWritePlaceFirmware(place int, bytes []byte)
	RunReadPlaceFirmware(place int)
}

type FirmwareInfo2 struct {
	Place                                  int
	Year, Month, Day, Hour, Minute, Second int
	Sensitivity, Serial, ProductType,
	Gas, Units, ScaleBegin, ScaleEnd string
	Values []string
}

type TempValues struct {
	Values []string
}

func (v FirmwareInfo2) GetFirmware() (data.Firmware, error) {
	z := data.Firmware{
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
	z.ScaleBegin, err = parseFloat(v.ScaleBegin)
	if err != nil {
		return z, merry.Appendf(err, "не верный формат значения начала шкалы: %s", v.ScaleBegin)
	}

	z.ScaleEnd, err = parseFloat(v.ScaleEnd)
	if err != nil {
		return z, merry.Appendf(err, "не верный формат значения конца шкалы: %s", v.ScaleEnd)
	}

	z.Fon, z.Sens = data.TableXY{}, data.TableXY{}

	if err = tempPoints(v.Values, z.Fon, z.Sens); err != nil {
		return z, merry.Append(err, "не удался расчёт температурных точек")
	}

	z.KSens20, err = parseFloat(v.Sensitivity)
	if err != nil {
		return z, merry.Appendf(err, "не верный формат значения коэффициента чувствиттельности: %s", v.Sensitivity)
	}

	z.ProductType = v.ProductType

	z.Serial, err = parseFloat(v.Serial)
	if err != nil {
		return z, merry.Appendf(err, "не верный формат значения серийного номера: %s", v.Serial)
	}
	return z, nil
}

func NewProductFirmware(f FirmwareRunner) *PlaceFirmware {
	return &PlaceFirmware{f}
}

func (x *PlaceFirmware) StoredFirmwareInfo(productID [1]int64, r *data.FirmwareInfo) error {

	var p data.Product
	if err := data.DB.SelectOneTo(&p, `WHERE product_id = ?`, productID[0]); err != nil {
		return err
	}
	if len(p.Firmware) == 0 {
		return merry.New("ЭХЯ не \"прошита\"")
	}
	if len(p.Firmware) < data.FirmwareSize {
		return merry.New("не верный формат \"прошивки\"")
	}
	*r = data.FirmwareBytes(p.Firmware).FirmwareInfo(p.Place)
	return nil
}

func (x *PlaceFirmware) CalculateFirmwareInfo(productID [1]int64, r *data.FirmwareInfo) (err error) {
	var p data.ProductInfo
	if err := data.DB.SelectOneTo(&p, `WHERE product_id = ?`, productID[0]); err != nil {
		return err
	}
	*r = p.FirmwareInfo()
	return nil
}

func (x *PlaceFirmware) TempPoints(v TempValues, r *data.TempPoints) error {
	fonM, sensM := data.TableXY{}, data.TableXY{}
	if err := tempPoints(v.Values, fonM, sensM); err != nil {
		return err
	}
	*r = data.NewTempPoints(fonM, sensM)
	return nil
}

func (x *PlaceFirmware) RunReadPlaceFirmware(place [1]int, _ *struct{}) error {
	x.f.RunReadPlaceFirmware(place[0])
	return nil
}

func (x *PlaceFirmware) SaveProductType(v struct{ X FirmwareInfo2 }, _ *struct{}) error {
	z, err := v.X.GetFirmware()
	if err != nil {
		return err
	}
	var p data.ProductType

	if err := data.DB.FindByPrimaryKeyTo(&p, v.X.ProductType); err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		if err := data.DB.FindByPrimaryKeyTo(&p, "035"); err != nil {
			return err
		}
	}
	p.ProductTypeName = z.ProductType
	p.Scale = z.ScaleEnd
	p.UnitsName = v.X.Units
	p.GasName = v.X.Gas

	if err := data.DB.Save(&p); err != nil {
		return err
	}

	for t, fon := range z.Fon {
		data.DBx.MustExec(`INSERT OR REPLACE INTO product_type_current (product_type_name, temperature, fon, current) VALUES (?, ?, 1, ?)`,
			p.ProductTypeName, t, fon)
	}
	for t, sens := range z.Sens {
		data.DBx.MustExec(`INSERT OR REPLACE INTO product_type_current (product_type_name, temperature, fon, current) VALUES (?, ?, 0, ?)`,
			p.ProductTypeName, t, sens)
	}
	return nil
}

func (x *PlaceFirmware) RunWritePlaceFirmware(v struct{ X FirmwareInfo2 }, _ *struct{}) error {
	z, err := v.X.GetFirmware()
	if err != nil {
		return err
	}
	x.f.RunWritePlaceFirmware(v.X.Place, z.Bytes())
	return nil
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
