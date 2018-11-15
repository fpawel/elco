package crud

import (
	"github.com/fpawel/elco/internal/crud/data"
	"github.com/jmoiron/sqlx"
	"gopkg.in/reform.v1"
	"sync"
)

type LastParty struct {
	mu   *sync.Mutex
	conn *sqlx.DB
	dbr  *reform.DB
}

func (x LastParty) Party() (data.PartyInfo, []data.ProductInfo) {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.party()
}

func (x LastParty) ProductAtPlace(place int) (product data.ProductInfo, err error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	party, _ := x.party()
	err = x.dbr.SelectOneTo(&product, "WHERE party_id = ? AND place = ?", party.PartyID, place)
	return
}

func (x LastParty) SetProductSerialAtPlace(place, serial int) (int64, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	party, _ := x.party()
	p := data.Product{PartyID: party.PartyID, Place: place}
	if err := x.dbr.SelectOneTo(&p, "WHERE party_id = ? AND place = ?", party.PartyID, place); err != reform.ErrNoRows {
		return 0, err
	}
	if err := x.dbr.Save(&p); err != nil {
		return 0, err
	}
	return p.ProductID, nil
}

func (x LastParty) ToggleProductProductionAtPlace(place int) error {
	x.mu.Lock()
	defer x.mu.Unlock()
	party, _ := x.party()
	p := data.Product{PartyID: party.PartyID, Place: place}
	if err := x.dbr.SelectOneTo(&p, "WHERE party_id = ? AND place = ?", party.PartyID, place); err != reform.ErrNoRows {
		return err
	}
	p.Production = !p.Production
	return x.dbr.Save(&p)
}

func (x LastParty) party() (data.PartyInfo, []data.ProductInfo) {
	var party data.PartyInfo
	if err := x.dbr.SelectOneTo(&party, `ORDER BY created_at DESC LIMIT 1;`); err != nil {
		panic(err)
	}
	return party, data.GetProductsByPartyID(x.dbr, party.PartyID)
}
