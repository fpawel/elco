package crud

import (
	"github.com/fpawel/elco/internal/crud/data"
	"github.com/jmoiron/sqlx"
	"gopkg.in/reform.v1"
	"reflect"
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
	var p data.Product
	if err := x.dbr.SelectOneTo(&p, "WHERE party_id = ? AND place = ?", party.PartyID, place); err != nil && err != reform.ErrNoRows {
		return 0, err
	}
	p.PartyID = party.PartyID
	p.Place = place
	p.Serial.Int64 = int64(serial)
	p.Serial.Valid = true

	if err := x.dbr.Save(&p); err != nil {
		return 0, err
	}
	return p.ProductID, nil
}

func (x LastParty) ToggleProductProductionAtPlace(place int) (int64, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	party, _ := x.party()
	p := data.Product{PartyID: party.PartyID, Place: place}
	empty1 := p

	if err := x.dbr.SelectOneTo(&p, "WHERE party_id = ? AND place = ?", party.PartyID, place); err != nil && err != reform.ErrNoRows {
		return 0, err
	}
	empty1.ProductID = p.ProductID
	empty2 := empty1
	empty2.Firmware = []byte{}

	p.Production = !p.Production
	if reflect.DeepEqual(p, empty1) || reflect.DeepEqual(p, empty2) {
		return 0, x.dbr.Delete(&p)
	}

	if err := x.dbr.Save(&p); err != nil {
		return 0, err
	}
	return p.ProductID, nil
}

func (x LastParty) party() (data.PartyInfo, []data.ProductInfo) {
	var party data.PartyInfo
	if err := x.dbr.SelectOneTo(&party, `ORDER BY created_at DESC LIMIT 1;`); err != nil {
		panic(err)
	}
	return party, data.GetProductsByPartyID(x.dbr, party.PartyID)
}
