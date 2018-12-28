package crud

import (
	"encoding/json"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/app"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/goutils/dbutils"
	"github.com/sirupsen/logrus"
	"gopkg.in/reform.v1"
	"io/ioutil"
	"time"
)

type PartiesCatalogue struct {
	dbContext
}

func (x PartiesCatalogue) LastParty() data.Party {
	return LastParty{x.dbContext}.Party()
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
			Note:            p.Note,
		})
	}
	if err != reform.ErrNoRows {
		panic(err)
	}
	return
}

func (x PartiesCatalogue) Party(partyID int64) (party data.Party, err error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	if err = x.dbr.FindOneTo(&party, "party_id", partyID); err != nil {
		return
	}
	party.Products = data.GetProductsInfoByPartyID(x.dbr, partyID)
	return

}

func (x PartiesCatalogue) ProductInfo(productID int64) (product data.ProductInfo) {
	x.mu.Lock()
	defer x.mu.Unlock()
	if err := x.dbr.FindOneTo(&product, "product_id", productID); err != nil {
		panic(err)
	}
	return
}

func (x PartiesCatalogue) ImportFromFile() (data.Party, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	if err := x.importFromFile(); err != nil {
		return data.Party{}, err
	}
	return data.MustLastParty(x.dbr), nil
}

func (x PartiesCatalogue) DeletePartyID(partyID int64) error {
	x.mu.Lock()
	defer x.mu.Unlock()
	_, err := x.dbx.Exec(`DELETE FROM party WHERE party_id = ?`, partyID)

	logrus.WithFields(logrus.Fields{
		"party_id": partyID,
		"result":   err,
	}).Warn("delete party id")

	data.EnsureParty(x.dbx)
	return err
}

func (x PartiesCatalogue) DeleteDay(year, month, day int) error {
	x.mu.Lock()
	defer x.mu.Unlock()
	_, err := x.dbx.Exec(`
DELETE FROM party 
WHERE 
      cast(strftime('%Y', DATETIME(created_at, '+3 hours')) AS INTEGER) = ? AND
      cast(strftime('%m', DATETIME(created_at, '+3 hours')) AS INTEGER) = ? AND
      cast(strftime('%d', DATETIME(created_at, '+3 hours')) AS INTEGER) = ?`, year, month, day)

	logrus.WithFields(logrus.Fields{
		"year":   year,
		"month":  month,
		"day":    day,
		"result": err,
	}).Warn("delete day")

	data.EnsureParty(x.dbx)
	return err
}

func (x PartiesCatalogue) DeleteMonth(year, month int) error {
	x.mu.Lock()
	defer x.mu.Unlock()
	_, err := x.dbx.Exec(`
DELETE FROM party 
WHERE 
      cast(strftime('%Y', DATETIME(created_at, '+3 hours')) AS INTEGER) = ? AND
      cast(strftime('%m', DATETIME(created_at, '+3 hours')) AS INTEGER) = ?`, year, month)

	logrus.WithFields(logrus.Fields{
		"year":   year,
		"month":  month,
		"result": err,
	}).Warn("delete month")

	data.EnsureParty(x.dbx)
	return err
}

func (x PartiesCatalogue) DeleteYear(year int) error {
	x.mu.Lock()
	defer x.mu.Unlock()
	_, err := x.dbx.Exec(`
DELETE FROM party 
WHERE cast(strftime('%Y', DATETIME(created_at, '+3 hours')) AS INTEGER) = ?`, year)

	logrus.WithFields(logrus.Fields{
		"year":   year,
		"result": err,
	}).Warn("delete year")

	data.EnsureParty(x.dbx)

	return err
}

func (x PartiesCatalogue) importFromFile() error {

	b, err := ioutil.ReadFile(importFileName())
	if err != nil {
		return err
	}
	var oldParty data.OldParty
	if err := json.Unmarshal(b, &oldParty); err != nil {
		return err
	}
	party, products := oldParty.Party()
	party.CreatedAt = time.Now()
	if err := x.dbr.Save(&party); err != nil {
		return err
	}
	for _, p := range products {
		p.PartyID = party.PartyID
		if err := x.dbr.Save(&p); err != nil {
			return merry.Appendf(err, "product: serial: %v place: %d",
				p.Serial, p.Place)
		}
	}
	return nil
}

func importFileName() string {
	return app.AppName.DataFileName("export-party.json")
}
