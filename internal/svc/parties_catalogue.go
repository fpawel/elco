package svc

import (
	"github.com/fpawel/elco/internal/crud"
	"github.com/fpawel/elco/internal/crud/data"
)

type PartiesCatalogue struct {
	c crud.PartiesCatalogue
}

func NewPartiesCatalogue(c crud.PartiesCatalogue) *PartiesCatalogue {
	return &PartiesCatalogue{c}
}

func (x *PartiesCatalogue) Years(_ struct{}, years *[]int) error {
	*years = x.c.Years()
	return nil
}

func (x *PartiesCatalogue) Months(r struct{ Year int }, months *[]int) error {
	*months = x.c.Months(r.Year)
	return nil
}

func (x *PartiesCatalogue) Days(r struct{ Year, Month int }, days *[]int) error {
	*days = x.c.Days(r.Year, r.Month)
	return nil
}

func (x *PartiesCatalogue) Parties(r struct{ Year, Month, Day int }, parties *[]data.Party) error {
	*parties = x.c.Parties(r.Year, r.Month, r.Day)
	return nil
}

func (x *PartiesCatalogue) Party(partyID [1]int64, r *Party) error {
	r.Party, r.Products = x.c.Party(partyID[0])
	return nil
}
