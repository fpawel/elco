package api

import (
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/data/chipmem"
	"strconv"
)

type PlaceFirmware struct {
	f FirmwareRunner
}

type FirmwareRunner interface {
	RunWritePlaceFirmware(placeDevice, placeProduct int, bytes []byte) error
	RunReadPlaceFirmware(place int)
}

type TempValues struct {
	Values []string
}

func NewProductFirmware(f FirmwareRunner) *PlaceFirmware {
	return &PlaceFirmware{f}
}

func (x *PlaceFirmware) StoredProductFirmware(productID [1]int64, r *chipmem.FirmwareInfo) error {

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
	*r = chipmem.Bytes(p.Firmware).FirmwareInfo()
	//for _, b := range p.Firmware {
	//	r.Bytes = append(r.Bytes, fmt.Sprintf("%02X", b))
	//}
	//if len(r.Bytes) == 0 {
	//	r.Bytes = []string{}
	//}
	return nil
}

func (x *PlaceFirmware) CalculatedProductFirmware(productID [1]int64, r *chipmem.FirmwareInfo) error {
	var p data.ProductInfo
	err := data.DB.SelectOneTo(&p, `WHERE product_id = ?`, productID[0])
	if err != nil {
		return err
	}
	*r = chipmem.ProductInfo{P: p}.FirmwareInfo()
	//r.Bytes, err = r.FirmwareInfo.CalculateBytes()
	return err
}

func (x *PlaceFirmware) CalculateTempPoints(v struct{ TempValues []string }, r *chipmem.TempPoints) error {
	fonM, sensM := chipmem.TableXY{}, chipmem.TableXY{}
	if err := chipmem.GetTempTables(v.TempValues, fonM, sensM); err != nil {
		return err
	}
	*r = chipmem.NewTempPoints(fonM, sensM)
	return nil
}

func (x *PlaceFirmware) GetFirmwareBytes(v struct {
	FirmwareInfo chipmem.FirmwareInfo
}, r *[]string) error {

	f, err := v.FirmwareInfo.GetFirmware()
	if err != nil {
		return err
	}
	for _, b := range f.Bytes() {
		*r = append(*r, fmt.Sprintf("%02X", b))
	}
	return nil
}

func (x *PlaceFirmware) SetFirmwareBytes(v struct {
	Bytes []string
}, r *chipmem.FirmwareInfo) error {
	if len(v.Bytes) == 0 || len(v.Bytes) < chipmem.FirmwareSize {
		return merry.New("out of range")
	}
	xs := chipmem.Bytes{}
	for i, s := range v.Bytes {
		if i >= chipmem.FirmwareSize {
			break
		}
		b, err := strconv.ParseUint("0x"+s, 0, 8)
		if err != nil {
			return merry.Appendf(err, "адрес %04X: %q", i, s)
		}
		xs = append(xs, byte(b))
	}
	*r = xs.FirmwareInfo()
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

	xs, err := data.GetProductTypeTempPoints(v.X.TempValues, v.X.ProductType)
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
	FirmwareInfo              chipmem.FirmwareInfo
	PlaceDevice, PlaceProduct int
}, _ *struct{}) error {
	z, err := arg.FirmwareInfo.GetFirmware()
	if err != nil {
		return err
	}
	return x.f.RunWritePlaceFirmware(arg.PlaceDevice, arg.PlaceProduct, z.Bytes())
}
