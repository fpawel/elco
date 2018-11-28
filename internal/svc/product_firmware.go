package svc

import (
	"github.com/fpawel/elco/internal/crud"
	"github.com/fpawel/elco/internal/firmware"
)

type ProductFirmware struct {
	c crud.ProductFirmware
}

func NewProductFirmware(c crud.ProductFirmware) *ProductFirmware {
	return &ProductFirmware{c}
}

func (x *ProductFirmware) Stored(productID [1]int64, r *firmware.FlashInfo) error {
	if b, err := x.c.Stored(productID[0]); err != nil {
		return err
	} else {
		*r = b.Info()
		return nil
	}
}

func (x *ProductFirmware) Calculated(productID [1]int64, r *firmware.FlashInfo) error {
	b, err := x.c.Calculated(productID[0])
	if b != nil {
		*r = b.Info()
	}
	return err
}
