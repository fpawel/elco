package api

import (
	"github.com/fpawel/elco/internal/data"
	"gopkg.in/reform.v1"
)

type ProductTypes struct {
	db *reform.DB
}

func NewProductTypes(db *reform.DB) *ProductTypes {
	return &ProductTypes{db}
}

func (x *ProductTypes) Names(_ struct{}, r *[]string) (err error) {
	*r, err = data.ListProductTypeNames(x.db)
	return nil
}

func (x *ProductTypes) Gases(_ struct{}, r *[]string) error {
	gases, err := data.ListGases(x.db)
	if err != nil {
		return err
	}
	for _, g := range gases {
		*r = append(*r, g.GasName)
	}
	return nil
}

func (x *ProductTypes) Units(_ struct{}, r *[]string) error {
	units, err := data.ListUnits(x.db)
	if err != nil {
		return err
	}
	for _, g := range units {
		*r = append(*r, g.UnitsName)
	}
	return nil
}
