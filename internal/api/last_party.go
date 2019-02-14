package api

import (
	"github.com/fpawel/elco/internal/data"
	"gopkg.in/reform.v1"
	"strings"
)

type LastParty struct {
	db *reform.DB
}

func NewLastParty(db *reform.DB) *LastParty {
	return &LastParty{db}
}

func (x *LastParty) Party(_ struct{}, r *Party) error {
	party, err := data.GetLastParty(x.db)
	if err != nil {
		return err
	}
	*r, err = makeParty(x.db, party, party.PartyID)
	return err
}

func (x *LastParty) SetProductSerialAtPlace(p [2]int, r *int64) (err error) {
	*r, err = data.UpdateProductAtPlace(x.db, p[0], func(product *data.Product) error {
		product.Serial.Int64 = int64(p[1])
		product.Serial.Valid = true
		return nil
	})
	return
}

func (x LastParty) ProductAtPlace(place [1]int, r *data.ProductInfo) error {
	partyID, err := data.GetLastPartyID(x.db)
	if err != nil {
		return err
	}
	return x.db.SelectOneTo(r, "WHERE party_id = ? AND place = ?", partyID, place)
}

func (x LastParty) ToggleProductProductionAtPlace(place [1]int, r *int64) (err error) {
	*r, err = data.UpdateProductAtPlace(x.db, place[0], func(p *data.Product) error {
		p.Production = !p.Production
		return nil
	})
	return
}

func (x LastParty) SetProductNoteAtPlace(p struct {
	Place int
	Note  string
}, r *int64) (err error) {
	*r, err = data.UpdateProductAtPlace(x.db, p.Place, func(product *data.Product) error {
		product.Note.String = strings.TrimSpace(p.Note)
		product.Note.Valid = len(product.Note.String) > 0
		return nil
	})
	return
}

func (x LastParty) SetProductTypeAtPlace(p struct {
	Place       int
	ProductType string
}, r *int64) (err error) {
	*r, err = data.UpdateProductAtPlace(x.db, p.Place, func(product *data.Product) error {
		product.ProductTypeName.String = strings.TrimSpace(p.ProductType)
		product.ProductTypeName.Valid = len(product.ProductTypeName.String) > 0
		return nil
	})
	return
}

func (x LastParty) SetPointsMethodAtPlace(p struct {
	Place        int
	PointsMethod int64
	Valid        bool
}, r *int64) (err error) {
	*r, err = data.UpdateProductAtPlace(x.db, p.Place, func(product *data.Product) error {
		product.PointsMethod.Int64 = p.PointsMethod
		product.PointsMethod.Valid = p.Valid
		return nil
	})
	return
}

func (x LastParty) DeleteProductAtPlace(place [1]int, _ *struct{}) (err error) {
	_, err = x.db.Exec(`
DELETE FROM product 
WHERE party_id IN (
  SELECT party.party_id 
  FROM party 
  ORDER BY created_at DESC 
  LIMIT 1) AND place = ?`, place[0])
	return
}

func (x LastParty) Export(_ struct{}, _ *struct{}) error {
	return data.ExportLastParty(x.db)
}

func (x *PartiesCatalogue) Import(_ struct{}, r *data.Party) (err error) {
	return data.ImportLastParty(x.db)
}
