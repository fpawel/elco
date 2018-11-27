package crud

import (
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/goutils/dbutils"
	"gopkg.in/reform.v1"
)

type PartiesCatalogue struct {
	dbContext
}

func (x PartiesCatalogue) Years() (years []int) {
	x.mu.Lock()
	defer x.mu.Unlock()
	dbutils.MustSelect(x.dbx, &years, `SELECT DISTINCT year FROM party_info ORDER BY year ASC;`)
	return
}

func (x PartiesCatalogue) Months(y int) (months []int) {
	x.mu.Lock()
	defer x.mu.Unlock()
	dbutils.MustSelect(x.dbx, &months,
		`SELECT DISTINCT month FROM party_info WHERE year = ? ORDER BY month ASC;`, y)
	return
}

func (x PartiesCatalogue) Days(year, month int) (days []int) {
	x.mu.Lock()
	defer x.mu.Unlock()
	dbutils.MustSelect(x.dbx, &days,
		`SELECT DISTINCT day FROM party_info WHERE year = ? AND month = ? ORDER BY day ASC;`,
		year, month)
	return
}

func (x PartiesCatalogue) Parties(year, month, day int) (parties []data.Party) {
	x.mu.Lock()
	defer x.mu.Unlock()

	rows, err := x.dbr.SelectRows(data.PartyInfoTable,
		"WHERE year = ? AND month = ? AND day = ?",
		year, month, day)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for {
		var p data.PartyInfo
		if err = x.dbr.NextRow(&p, rows); err != nil {
			break
		}
		parties = append(parties, data.Party{
			PartyID:         p.PartyID,
			CreatedAt:       p.CreatedAt,
			UpdatedAt:       p.UpdatedAt,
			ProductTypeName: p.ProductTypeName,
			Concentration1:  p.Concentration1,
			Concentration2:  p.Concentration2,
			Concentration3:  p.Concentration3,
			Last:            p.Last,
		})
	}
	if err != reform.ErrNoRows {
		panic(err)
	}
	return
}

func (x PartiesCatalogue) Party(partyID int64) (party data.Party) {
	x.mu.Lock()
	defer x.mu.Unlock()
	if err := x.dbr.FindOneTo(&party, "party_id", partyID); err != nil {
		panic(err)
	}
	party.Products = data.GetProductsByPartyID(x.dbr, partyID)
	return

}
