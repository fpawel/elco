package svc

import "github.com/fpawel/elco/internal/crud"

type ProductTypes struct {
	c crud.ProductTypes
}

func NewProductTypes(c crud.ProductTypes) *ProductTypes {
	return &ProductTypes{c}
}

func (x *ProductTypes) Names(_ struct{}, r *[]string) error {
	*r = x.c.ListProductTypesNames()
	return nil
}

func (x *ProductTypes) Gases(_ struct{}, r *[]string) error {
	for _, g := range x.c.ListGases() {
		*r = append(*r, g.GasName)
	}
	return nil
}

func (x *ProductTypes) Units(_ struct{}, r *[]string) error {
	for _, g := range x.c.ListUnits() {
		*r = append(*r, g.UnitsName)
	}
	return nil
}
