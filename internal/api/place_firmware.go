package api

import (
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/data/chipmem"
	"github.com/pkg/errors"
	"strings"
)

type PlaceFirmware struct {
	f FirmwareRunner
}

type FirmwareRunner interface {
	RunWritePlaceFirmware(placeDevice, placeProduct int, bytes []byte) error
	RunReadPlaceFirmware(place int)
}

type Firmware struct {
	FirmwareInfo chipmem.FirmwareInfo
	Bytes        []string
}

type TempValues struct {
	Values []string
}

func NewProductFirmware(f FirmwareRunner) *PlaceFirmware {
	return &PlaceFirmware{f}
}

func (x *PlaceFirmware) StoredFirmware(productID [1]int64, r *Firmware) error {

	var p data.Product
	if err := data.DB.SelectOneTo(&p, `WHERE product_id = ?`, productID[0]); err != nil {
		return err
	}
	if len(p.Firmware) == 0 {
		return merry.New("ЭХЯ не \"прошита\"")
	}
	if len(p.Firmware) < chipmem.FirmwareSize {
		return merry.New("не верный формат \"прошивки\"")
	}
	r.FirmwareInfo = chipmem.Bytes(p.Firmware).FirmwareInfo(p.Place)
	for _, b := range p.Firmware {
		r.Bytes = append(r.Bytes, fmt.Sprintf("%02X", b))
	}
	if len(r.Bytes) == 0 {
		r.Bytes = []string{}
	}
	return nil
}

func (x *PlaceFirmware) CalculateFirmware(productID [1]int64, r *Firmware) error {
	var p data.ProductInfo
	if err := data.DB.SelectOneTo(&p, `WHERE product_id = ?`, productID[0]); err != nil {
		return err
	}
	r.FirmwareInfo = chipmem.ProductInfo{P: p}.FirmwareInfo()
	firmware, err := r.FirmwareInfo.GetFirmware()
	if err != nil {
		return err
	}
	for _, b := range firmware.Bytes() {
		r.Bytes = append(r.Bytes, fmt.Sprintf("%02X", b))
	}
	return nil
}

func (x *PlaceFirmware) TempPoints(v TempValues, r *chipmem.TempPoints) error {
	fonM, sensM := data.TableXY{}, data.TableXY{}
	if err := tempPoints(v.Values, fonM, sensM); err != nil {
		return err
	}
	*r = chipmem.NewTempPoints(fonM, sensM)
	return nil
}

type ProductType2 struct {
	data.ProductType
	Currents   [][3]string
	TempPoints chipmem.TempPoints
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
		if x.Fon != nil {
			fonM[x.Temperature] = *x.Fon
		}
		if x.Sens != nil {
			sensM[x.Temperature] = *x.Sens
		}
		r.Currents = append(r.Currents,
			[3]string{
				formatFloat(x.Temperature, -1),
				formatFloatPtr(x.Fon, -1),
				formatFloatPtr(x.Sens, -1),
			})
	}
	r.TempPoints = chipmem.NewTempPoints(fonM, sensM)
	return nil
}

func (x *PlaceFirmware) RunReadPlaceFirmware(place [1]int, _ *struct{}) error {
	x.f.RunReadPlaceFirmware(place[0])
	return nil
}

func (x *PlaceFirmware) SaveProductType(v struct{ X chipmem.FirmwareInfo }, _ *struct{}) error {

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

	xs, err := tempPointsProductType(v.X.TempValues, v.X.ProductType)
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

func (x *PlaceFirmware) RunWritePlaceFirmware(arg struct {
	FirmwareInfo chipmem.FirmwareInfo
	PlaceDevice  int
}, _ *struct{}) error {
	z, err := arg.FirmwareInfo.GetFirmware()
	if err != nil {
		return err
	}
	return x.f.RunWritePlaceFirmware(arg.PlaceDevice, arg.FirmwareInfo.Place, z.Bytes())
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

func tempPointsProductType(values []string, productTypeName string) ([]data.ProductTypeTempPoint, error) {
	if len(values)%3 != 0 {
		return nil, errors.New("sequence length is not a multiple of three")
	}

	var xs []data.ProductTypeTempPoint

	for n := 0; n < len(values); n += 3 {
		strT := strings.TrimSpace(values[n+0])
		if len(strT) == 0 {
			continue
		}

		r := data.ProductTypeTempPoint{ProductTypeName: productTypeName}
		var err error

		r.Temperature, err = parseFloat(values[n])
		if err != nil {
			return nil, merry.Appendf(err, "строка %d", n)
		}

		strI := strings.TrimSpace(values[n+1])
		r.Fon, err = parseFloatPtr(strI)
		if err != nil {
			return nil, merry.Appendf(err, "строка %d", n)
		}

		strS := strings.TrimSpace(values[n+2])
		r.Sens, err = parseFloatPtr(strS)
		if err != nil {
			return nil, merry.Appendf(err, "строка %d", n)
		}
		xs = append(xs, r)
	}
	return xs, nil
}
