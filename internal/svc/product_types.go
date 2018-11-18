package svc

import "github.com/fpawel/elco/internal/crud"

type ProductTypes struct {
	c crud.ProductTypes
}

func NewProductTypes(c crud.ProductTypes) *ProductTypes {
	return &ProductTypes{c}
}

func (x *ProductTypes) Names(_ struct{}, r *[]string) error {
	*r = x.c.Names()
	return nil
}
