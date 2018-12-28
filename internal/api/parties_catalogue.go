package api

import (
	"github.com/fpawel/elco/internal/crud"
	"github.com/fpawel/elco/internal/data"
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

func (x *PartiesCatalogue) Party(a [1]int64, r *data.Party) (err error) {
	*r, err = x.c.Party(a[0])
	return
}

func (x *PartiesCatalogue) NewParty(_ struct{}, r *data.Party) error {
	*r = x.c.NewParty()
	return nil
}

func (x *PartiesCatalogue) ImportFromFile(_ struct{}, r *data.Party) (err error) {
	*r, err = x.c.ImportFromFile()
	return
}

func (x *PartiesCatalogue) DeletePartyID(partyID [1]int64, _ *struct{}) error {
	return x.c.DeletePartyID(partyID[0])
}

func (x *PartiesCatalogue) DeleteDay(v [3]int, _ *struct{}) error {
	return x.c.DeleteDay(v[0], v[1], v[2])
}

func (x *PartiesCatalogue) DeleteMonth(v [2]int, _ *struct{}) error {
	return x.c.DeleteMonth(v[0], v[1])
}

func (x *PartiesCatalogue) DeleteYear(v [1]int, _ *struct{}) error {
	return x.c.DeleteYear(v[0])
}
