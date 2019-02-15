package api

import (
	"database/sql"
	"github.com/fpawel/elco/internal/data"
	"gopkg.in/reform.v1"
)

type PartiesCatalogue struct {
	db *reform.DB
}

func NewPartiesCatalogue(db *reform.DB) *PartiesCatalogue {
	return &PartiesCatalogue{db}
}

func (x *PartiesCatalogue) Years(_ struct{}, years *[]int) error {
	rows, err := x.db.Query(`SELECT DISTINCT year FROM party_info ORDER BY year ASC;`)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var n int
		err := rows.Scan(&n)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil
			}
			return err
		}
		*years = append(*years, n)
	}
	return nil
}

func (x *PartiesCatalogue) Months(r struct{ Year int }, months *[]int) error {
	rows, err := x.db.Query(`
SELECT DISTINCT month FROM party_info WHERE year = ? ORDER BY month ASC;`, r.Year)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var n int
		err := rows.Scan(&n)
		if err != nil {
			if err != sql.ErrNoRows {
				return nil
			}
			return err
		}
		*months = append(*months, n)
	}
	return nil
}

func (x *PartiesCatalogue) Days(r struct{ Year, Month int }, days *[]int) error {
	rows, err := x.db.Query(`
SELECT DISTINCT day FROM party_info WHERE year = ? AND month = ? ORDER BY day ASC;`,
		r.Year, r.Month)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var n int
		err := rows.Scan(&n)
		if err != nil {
			if err != sql.ErrNoRows {
				return nil
			}
			return err
		}
		*days = append(*days, n)
	}
	return nil
}

func (x *PartiesCatalogue) Parties(r struct{ Year, Month, Day int }, parties *[]data.Party) error {
	xs, err := x.db.SelectAllFrom(data.PartyTable, `WHERE 
cast(strftime('%Y', DATETIME(created_at, '+3 hours')) AS INTEGER) = ? AND 
cast(strftime('%m', DATETIME(created_at, '+3 hours')) AS INTEGER) = ? AND 
cast(strftime('%d', DATETIME(created_at, '+3 hours')) AS INTEGER) = ?`,
		r.Year, r.Month, r.Day)
	if err != nil {
		return err
	}

	lastPartyID, err := data.GetLastPartyID(x.db)
	if err != nil {
		return err
	}
	for _, y := range xs {
		party := y.(*data.Party)
		party.Last = party.PartyID == lastPartyID
		*parties = append(*parties, *party)
	}

	return nil
}

func (x *PartiesCatalogue) Party(a [1]int64, r *data.Party) error {
	if err := x.db.FindByPrimaryKeyTo(r, a[0]); err != nil {
		return err
	}
	if err := data.GetPartyProducts(x.db, r); err != nil {
		return err
	}
	return data.GetPartyIsLast(x.db, r)
}

func (x *PartiesCatalogue) NewParty(_ struct{}, r *data.Party) error {
	newPartyID, err := data.CreateNewParty(x.db)
	if err != nil {
		return err
	}
	return x.db.FindByPrimaryKeyTo(r, newPartyID)
}

func (x *PartiesCatalogue) DeletePartyID(partyID [1]int64, _ *struct{}) error {
	if _, err := x.db.Exec(`DELETE FROM party WHERE party_id = ?`, partyID[0]); err != nil {
		return err
	}
	_, err := data.GetLastPartyID(x.db)
	return err
}

func (x *PartiesCatalogue) DeleteDay(v [3]int, _ *struct{}) error {
	if _, err := x.db.Exec(`
DELETE FROM party 
WHERE 
      cast(strftime('%Y', DATETIME(created_at, '+3 hours')) AS INTEGER) = ? AND
      cast(strftime('%m', DATETIME(created_at, '+3 hours')) AS INTEGER) = ? AND
      cast(strftime('%d', DATETIME(created_at, '+3 hours')) AS INTEGER) = ?`, v[0], v[1], v[2]); err != nil {
		return err
	}
	_, err := data.GetLastPartyID(x.db)
	return err
}

func (x *PartiesCatalogue) DeleteMonth(v [2]int, _ *struct{}) error {
	if _, err := x.db.Exec(`
DELETE FROM party 
WHERE 
      cast(strftime('%Y', DATETIME(created_at, '+3 hours')) AS INTEGER) = ? AND
      cast(strftime('%m', DATETIME(created_at, '+3 hours')) AS INTEGER) = ?`, v[0], v[1]); err != nil {
		return err
	}
	_, err := data.GetLastPartyID(x.db)
	return err
}

func (x *PartiesCatalogue) DeleteYear(v [1]int, _ *struct{}) error {
	if _, err := x.db.Exec(`
DELETE FROM party 
WHERE cast(strftime('%Y', DATETIME(created_at, '+3 hours')) AS INTEGER) = ?`, v[0]); err != nil {
		return err
	}
	_, err := data.GetLastPartyID(x.db)
	return err
}
