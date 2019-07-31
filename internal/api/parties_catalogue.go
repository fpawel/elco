package api

import (
	"database/sql"
	"fmt"
	"github.com/fpawel/elco/internal/data"
	"strconv"
)

type PartiesCatalogueSvc struct {
}

type YearMonth struct {
	Year  int `db:"year"`
	Month int `db:"month"`
}

type Party2 struct {
	PartyID         int            `db:"party_id"`
	Day             int            `db:"day"`
	ProductTypeName string         `db:"product_type_name"`
	Note            sql.NullString `db:"note"`
	Last            bool           `db:"last"`
}

func (_ *PartiesCatalogueSvc) YearsMonths(_ struct{}, r *[]YearMonth) error {
	if err := data.DBx.Select(r, `SELECT DISTINCT year, month FROM party_info ORDER BY year DESC, month DESC`); err != nil {
		panic(err)
	}
	return nil
}

func (_ *PartiesCatalogueSvc) PartiesOfYearMonth(x YearMonth, r *[]Party2) error {
	if err := data.DBx.Select(r, `
SELECT cast(strftime('%d', DATETIME(created_at, '+3 hours')) AS INTEGER)  AS day, 
       party_id, note, product_type_name,
       party_id = (SELECT party_id FROM party ORDER BY created_at DESC LIMIT 1)  AS last
FROM party
WHERE cast(strftime('%Y', DATETIME(created_at, '+3 hours')) AS INTEGER) = ?
  AND cast(strftime('%m', DATETIME(created_at, '+3 hours')) AS INTEGER) = ?
ORDER BY created_at`, x.Year, x.Month); err != nil {
		panic(err)
	}
	return nil
}

func (x *PartiesCatalogueSvc) Years(_ struct{}, years *[]int) error {
	return data.DBx.Select(years, `SELECT DISTINCT year FROM party_info ORDER BY year;`)
}

func (x *PartiesCatalogueSvc) Months(r struct{ Year int }, months *[]int) error {
	return data.DBx.Select(months,
		`SELECT DISTINCT month FROM party_info WHERE year = ? ORDER BY month;`,
		r.Year)
}

func (x *PartiesCatalogueSvc) Days(r struct{ Year, Month int }, days *[]int) error {
	return data.DBx.Select(days,
		`SELECT DISTINCT day FROM party_info WHERE year = ? AND month = ? ORDER BY day;`,
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
	*r, err = data.GetParty(a[0], data.WithAllProducts)
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

func (x *PartiesCatalogueSvc) SetProductsProduction(v struct {
	ProductIDs []int64
	Production bool
}, _ *struct{}) error {
	var s string
	for i, productID := range v.ProductIDs {
		if i > 0 {
			s += ","
		}
		s += strconv.FormatInt(productID, 10)
	}
	query := fmt.Sprintf(
		"UPDATE product SET production = ? WHERE product_id IN (%s)", s)

	data.DBx.MustExec(query, v.Production)

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
