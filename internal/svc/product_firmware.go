package svc

import (
	"github.com/fpawel/elco/internal/crud"
	"github.com/fpawel/elco/internal/firmware"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type ProductFirmware struct {
	c crud.ProductFirmware
}

type TempValues struct {
	Values []string
}

func NewProductFirmware(c crud.ProductFirmware) *ProductFirmware {
	return &ProductFirmware{c}
}

func (x *ProductFirmware) Stored(productID [1]int64, r *firmware.ProductFirmwareInfo) error {
	if b, err := x.c.Stored(productID[0]); err != nil {
		return err
	} else {
		*r = b.Info()
		return nil
	}
}

func (x *ProductFirmware) Calculate(productID [1]int64, r *firmware.ProductFirmwareInfo) (err error) {
	*r, err = x.c.Calculate(productID[0])
	return
}

func (x *ProductFirmware) CalculateTempPoints(v TempValues, r *firmware.TempPoints) (err error) {

	if len(v.Values) % 3 != 0 {
		return errors.New("sequence length is not a multiple of three")
	}

	fonM, sensM := firmware.M{}, firmware.M{}
	for n:=0 ; n < len(v.Values); n+=3 {
		strT := strings.TrimSpace(v.Values[n+0])
		if len(strT) == 0 {
			continue
		}

		var t float64
		t,err = strconv.ParseFloat(v.Values[n],64)
		if err != nil {
			return errors.Wrapf(err, "строка %d", n)
		}
		strI := strings.TrimSpace(v.Values[n+1])
		if len(strI) > 0 {
			var i float64
			i,err =  strconv.ParseFloat(strI,64)
			if err != nil {
				return errors.Wrapf(err, "строка %d", n)
			}
			fonM[t] = i
		}
		strS := strings.TrimSpace(v.Values[n+2])
		if len(strS) > 0 {
			var k float64
			k,err =  strconv.ParseFloat(strS,64)
			if err != nil {
				return errors.Wrapf(err, "строка %d", n)
			}
			sensM[t] = k
		}
	}
	*r = firmware.CalculateTempPoints(fonM,sensM)
	return
}
