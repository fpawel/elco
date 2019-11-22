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
	SensitivityProduct, SensitivityLab73,
	Fon20,
	Serial, ProductType,
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

type ProductType2 struct {
	data.ProductType
	Currents   [][3]string
	TempPoints data.TempPoints
}

func (x *PlaceFirmware) GetProductType(productTypeName [1]string, r *ProductType2) error {
	err := data.DB.FindByPrimaryKeyTo(&r.ProductType, productTypeName[0])
	if err == sql.ErrNoRows {
		return err
	}
	if err != nil {
		return err
	}
	r.Currents = [][3]string{}

	fonM, sensM := data.TableXY{}, data.TableXY{}
	xs := data.FetchProductTypeTempPoints(productTypeName[0])
	for _, x := range xs {
		fonM[x.T] = x.Fon
		sensM[x.T] = x.Sens
		r.Currents = append(r.Currents,
			[3]string{formatFloat(x.T, -1), formatFloat(x.Fon, -1), formatFloat(x.Sens, -1)})
	}
	r.TempPoints = data.NewTempPoints(fonM, sensM)
	return nil
}

func (x *PlaceFirmware) RunReadPlaceFirmware(place [1]int, _ *struct{}) error {
	x.f.RunReadPlaceFirmware(place[0])
	return nil
}

func (x *PlaceFirmware) SaveProductType(v struct{ X FirmwareInfo2 }, _ *struct{}) error {

	var p data.ProductType

	if err := data.DB.FindByPrimaryKeyTo(&p, v.X.ProductType); err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		if err := data.DB.FindByPrimaryKeyTo(&p, "035"); err != nil {
			return err
		}
	}
	p.ProductTypeName = v.X.ProductType
	p.UnitsName = v.X.Units
	p.GasName = v.X.Gas

	if value, err := parseFloat(v.X.ScaleEnd); err == nil {
		p.Scale = value
	} else {
		return merry.Appendf(err, "не верный формат значения кконца шкалы: %s", v.X.ScaleEnd)
	}
	if v, err := parseFloat(v.X.SensitivityProduct); err == nil {
		p.KSens20 = sql.NullFloat64{v, true}
	} else {
		p.KSens20 = sql.NullFloat64{}
	}

	if v, err := parseFloat(v.X.Fon20); err == nil {
		p.Fon20 = sql.NullFloat64{v, true}
	} else {
		p.Fon20 = sql.NullFloat64{}
	}

	if err := data.DB.Save(&p); err != nil {
		return err
	}

	xs, err := tempPointsProductType(v.X.Values, v.X.ProductType)
	if err != nil {
		return err
	}

	data.DBx.MustExec(`DELETE FROM product_type_current WHERE product_type_name=?`, v.X.ProductType)
	for _, v := range xs {
		_, err := data.DBx.NamedExec(`
INSERT INTO product_type_current (product_type_name, temperature, fon, sens) 
VALUES (:product_type_name, :temperature, :fon, :sens)`, v)
		if err != nil {
			return err
		}
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

type tempPointProductType struct {
	T float64 `db:"temperature"`
	I float64 `db:"fon"`
	S float64 `db:"sens"`
	P string  `db:"product_type_name"`
}

func tempPointsProductType(values []string, productTypeName string) ([]tempPointProductType, error) {
	if len(values)%3 != 0 {
		return nil, errors.New("sequence length is not a multiple of three")
	}

	var xs []tempPointProductType

	for n := 0; n < len(values); n += 3 {
		strT := strings.TrimSpace(values[n+0])
		if len(strT) == 0 {
			continue
		}

		r := tempPointProductType{P: productTypeName}
		var err error

		r.T, err = parseFloat(values[n])
		if err != nil {
			return nil, merry.Appendf(err, "строка %d", n)
		}

		strI := strings.TrimSpace(values[n+1])
		r.I, err = parseFloat(strI)
		if err != nil {
			return nil, merry.Appendf(err, "строка %d", n)
		}

		strS := strings.TrimSpace(values[n+2])
		r.S, err = parseFloat(strS)
		if err != nil {
			return nil, merry.Appendf(err, "строка %d", n)
		}
		xs = append(xs, r)
	}
	return xs, nil
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.Replace(s, ",", ".", -1), 64)
}

func formatNullFloat64K(v sql.NullFloat64, k float64, precision int) string {
	if v.Valid {
		return formatFloat(v.Float64*k, precision)
	}
	return ""
}

func formatFloat(v float64, precision int) string {
	return strconv.FormatFloat(v, 'f', precision, 64)
}
