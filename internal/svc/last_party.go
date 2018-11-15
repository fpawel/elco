package svc

import (
	"github.com/fpawel/elco/internal/crud"
	"github.com/fpawel/elco/internal/crud/data"
)

type LastParty struct {
	c crud.LastParty
}

func NewLastParty(c crud.LastParty) *LastParty {
	return &LastParty{c}
}

func (x *LastParty) Party(_ struct{}, r *Party) error {
	r.PartyInfo, r.Products = x.c.Party()
	return nil
}

func (x *LastParty) SetProductSerialAtPlace(p [2]int, r *int64) (err error) {
	*r, err = x.c.SetProductSerialAtPlace(p[0], p[1])
	return
}

func (x LastParty) ProductAtPlace(place [1]int, r *data.ProductInfo) (err error) {
	*r, err = x.c.ProductAtPlace(place[0])
	return
}

func (x LastParty) ToggleProductProductionAtPlace(place [1]int, r *int64) (err error) {
	*r, err = x.c.ToggleProductProductionAtPlace(place[0])
	return
}
