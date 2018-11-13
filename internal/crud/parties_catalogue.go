package crud

import (
	"github.com/fpawel/elco/internal/crud/data"
	"github.com/fpawel/goutils/dbutils"
	"github.com/jmoiron/sqlx"
	"gopkg.in/reform.v1"
	"sync"
)

type PartiesCatalogue struct {
	mu   *sync.Mutex
	conn *sqlx.DB
	dbr  *reform.DB
}

func (x PartiesCatalogue) Years() (years []int) {
	x.mu.Lock()
	defer x.mu.Unlock()
	dbutils.MustSelect(x.conn, &years, `SELECT DISTINCT year FROM party_info;`)
	return
}

func (x PartiesCatalogue) Months(y int) (months []int) {
	x.mu.Lock()
	defer x.mu.Unlock()
	dbutils.MustSelect(x.conn, &months,
		`SELECT DISTINCT month FROM party_info WHERE year = ?;`, y)
	return
}

func (x PartiesCatalogue) Days(year, month int) (days []int) {
	x.mu.Lock()
	defer x.mu.Unlock()
	dbutils.MustSelect(x.conn, &days,
		`
SELECT DISTINCT day 
FROM bucket_time 
WHERE year = ? AND month = ?;`,
		year, month)
	return
}

func (x PartiesCatalogue) Parties(year, month, day int) (parties []data.Party) {
	x.mu.Lock()
	defer x.mu.Unlock()

	rows, err := x.dbr.SelectRows(data.ProductInfoTable,
		"WHERE year = ? AND month = ? AND day = ?",
		year, month, day)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for {
		var party data.PartyInfo
		if err = x.dbr.NextRow(&party, rows); err != nil {
			break
		}
		parties = append(parties, party)
	}
	if err != reform.ErrNoRows {
		panic(err)
	}
	return
}
func (x PartiesCatalogue) Party(partyID int64) (party data.PartyInfo, products []data.ProductInfo) {
	x.mu.Lock()
	defer x.mu.Unlock()
	if err := x.dbr.FindOneTo(&party, "party_id", partyID); err != nil {
		panic(err)
	}

	rows, err := x.dbr.SelectRows(data.ProductInfoTable, "WHERE party_id = ? ORDER BY place",
		partyID)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for {
		var product data.ProductInfo
		if err = x.dbr.NextRow(&product, rows); err != nil {
			break
		}
		products = append(products, product)
	}
	if err != reform.ErrNoRows {
		panic(err)
	}

	return
}
