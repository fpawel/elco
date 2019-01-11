package api

import (
	"github.com/fpawel/elco/internal/crud"
	"github.com/fpawel/elco/internal/data"
)

type LastParty struct {
	c crud.LastParty
}

func NewLastParty(c crud.LastParty) *LastParty {
	return &LastParty{c}
}

func (x *LastParty) Party(_ struct{}, r *data.Party) error {
	*r = x.c.Party()
	return nil
}

func (x *LastParty) SetProductSerialAtPlace(p [2]int, r *int64) (err error) {
	*r, err = x.c.SetProductSerialAtPlace(p[0], p[1])
	return
}

func (x LastParty) ProductAtPlace(place [1]int, r *data.ProductInfo) (err error) {
	*r, err = x.c.GetProductInfoAtPlace(place[0])
	return
}

func (x LastParty) ToggleProductProductionAtPlace(place [1]int, r *int64) (err error) {
	*r, err = x.c.ToggleProductProductionAtPlace(place[0])
	return
}

func (x LastParty) SetProductNoteAtPlace(p struct {
	Place int
	Note  string
}, r *int64) (err error) {
	*r, err = x.c.SetProductNoteAtPlace(p.Place, p.Note)
	return
}

func (x LastParty) SetProductTypeAtPlace(p struct {
	Place       int
	ProductType string
}, r *int64) (err error) {
	*r, err = x.c.SetProductTypeAtPlace(p.Place, p.ProductType)
	return
}

func (x LastParty) SetPointsMethodAtPlace(p struct {
	Place        int
	PointsMethod int64
	Valid        bool
}, r *int64) (err error) {
	*r, err = x.c.SetPointsMethodAtPlace(p.Place, p.PointsMethod, p.Valid)
	return
}

func (x LastParty) DeleteProductAtPlace(place [1]int, _ *struct{}) (err error) {
	err = x.c.DeleteProductAtPlace(place[0])
	return
}

func (x LastParty) ExportToFile(_ struct{}, _ *struct{}) error {
	return x.c.ExportToFile()
}
