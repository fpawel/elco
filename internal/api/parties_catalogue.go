package api

import (
	"database/sql"
	"fmt"
	"github.com/fpawel/elco/internal/data"
	"time"
)

type PartiesCatalogueSvc struct {
}

func (x *PartiesCatalogueSvc) Years(_ struct{}, years *[]int) error {
	return data.DBx.Select(years, `SELECT DISTINCT year FROM party_info ORDER BY year ASC;`)
}

func (x *PartiesCatalogueSvc) Months(r struct{ Year int }, months *[]int) error {
	return data.DBx.Select(months,
		`SELECT DISTINCT month FROM party_info WHERE year = ? ORDER BY month ASC;`,
		r.Year)
}

func (x *PartiesCatalogueSvc) Days(r struct{ Year, Month int }, days *[]int) error {
	return data.DBx.Select(days,
		`SELECT DISTINCT day FROM party_info WHERE year = ? AND month = ? ORDER BY day ASC;`,
		r.Year, r.Month)
}

func (x *PartiesCatalogueSvc) Parties(r struct{ Year, Month, Day int }, parties *[]data.Party) error {
	xs, err := data.DB.SelectAllFrom(data.PartyTable, `WHERE 
cast(strftime('%Y', DATETIME(created_at, '+3 hours')) AS INTEGER) = ? AND 
cast(strftime('%m', DATETIME(created_at, '+3 hours')) AS INTEGER) = ? AND 
cast(strftime('%d', DATETIME(created_at, '+3 hours')) AS INTEGER) = ?`,
		r.Year, r.Month, r.Day)
	if err != nil {
		return err
	}

	lastPartyID := data.GetLastPartyID()
	for _, y := range xs {
		party := y.(*data.Party)
		party.Last = party.PartyID == lastPartyID
		*parties = append(*parties, *party)
	}

	return nil
}

func (x *PartiesCatalogueSvc) Party(a [1]int64, r *data.Party) (err error) {
	*r, err = data.GetParty(a[0], data.WithProducts)
	return
}

func (x *PartiesCatalogueSvc) NewParty(_ struct{}, r *data.Party) error {
	newPartyID := data.CreateNewParty()
	return data.DB.FindByPrimaryKeyTo(r, newPartyID)
}

func (x *PartiesCatalogueSvc) DeletePartyID(partyID [1]int64, _ *struct{}) error {
	if _, err := data.DB.Exec(`DELETE FROM party WHERE party_id = ?`, partyID[0]); err != nil {
		return err
	}
	return nil
}

func (x *PartiesCatalogueSvc) DeleteDay(v [3]int, _ *struct{}) error {
	if _, err := data.DB.Exec(`
DELETE FROM party 
WHERE 
      cast(strftime('%Y', DATETIME(created_at, '+3 hours')) AS INTEGER) = ? AND
      cast(strftime('%m', DATETIME(created_at, '+3 hours')) AS INTEGER) = ? AND
      cast(strftime('%d', DATETIME(created_at, '+3 hours')) AS INTEGER) = ?`, v[0], v[1], v[2]); err != nil {
		return err
	}
	return nil
}

func (x *PartiesCatalogueSvc) DeleteMonth(v [2]int, _ *struct{}) error {
	if _, err := data.DB.Exec(`
DELETE FROM party 
WHERE 
      cast(strftime('%Y', DATETIME(created_at, '+3 hours')) AS INTEGER) = ? AND
      cast(strftime('%m', DATETIME(created_at, '+3 hours')) AS INTEGER) = ?`, v[0], v[1]); err != nil {
		return err
	}
	return nil
}

func (x *PartiesCatalogueSvc) DeleteYear(v [1]int, _ *struct{}) error {
	if _, err := data.DB.Exec(`
DELETE FROM party 
WHERE cast(strftime('%Y', DATETIME(created_at, '+3 hours')) AS INTEGER) = ?`, v[0]); err != nil {
		return err
	}
	return nil
}

func (x *PartiesCatalogueSvc) CopyParty(partyID [1]int64, party *data.Party) (err error) {

	if err = data.DB.FindByPrimaryKeyTo(party, partyID[0]); err != nil {
		return err
	}
	s := fmt.Sprintf("Копия партии %d %s", partyID[0],
		party.CreatedAt.Format("2006.01.02"))
	if party.Note.Valid {
		s += ", " + party.Note.String
	}
	party.PartyID = 0
	party.Note = sql.NullString{s, true}
	party.CreatedAt = time.Now()
	if err = data.DB.Save(party); err != nil {
		return err
	}

	xsProducts, err := data.DB.SelectAllFrom(data.ProductTable, "WHERE party_id = ?", partyID[0])
	if err != nil {
		return err
	}
	for _, p := range xsProducts {
		product := p.(*data.Product)
		product.ProductID = 0
		product.PartyID = party.PartyID
		if err = data.DB.Save(product); err != nil {
			return err
		}
	}
	party.Products = data.GetProductsInfoWithPartyID(party.PartyID)
	return nil
}

func (x *PartiesCatalogueSvc) SetProductProduction(v struct {
	ProductID  int64
	Production bool
}, _ *struct{}) error {
	var p data.Product
	if err := data.DB.FindByPrimaryKeyTo(&p, v.ProductID); err != nil {
		return err
	}
	p.Production = v.Production
	if err := data.DB.Save(&p); err != nil {
		return err
	}
	return nil
}

func (x *PartiesCatalogueSvc) ToggleProductProduction(productID [1]int64, _ *struct{}) error {
	var p data.Product
	if err := data.DB.FindByPrimaryKeyTo(&p, productID[0]); err != nil {
		return err
	}
	p.Production = !p.Production
	if err := data.DB.Save(&p); err != nil {
		return err
	}
	return nil
}
