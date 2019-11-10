package api

import (
	"github.com/fpawel/elco/internal/data"
)

type ProductTypesSvc struct {
}

func (x *ProductTypesSvc) Names(_ struct{}, r *[]string) (err error) {
	*r = data.ProductTypeNames()
	return nil
}

func (x *ProductTypesSvc) Gases(_ struct{}, r *[]string) error {
	for _, g := range data.ListGases() {
		*r = append(*r, g.GasName)
	}
	return nil
}

func (x *ProductTypesSvc) Units(_ struct{}, r *[]string) error {
	for _, g := range data.ListUnits() {
		*r = append(*r, g.UnitsName)
	}
	return nil
}
