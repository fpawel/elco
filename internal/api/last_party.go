package api

import (
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/pdf"
	"github.com/jmoiron/sqlx"
	"gopkg.in/reform.v1"
	"strings"
)

type LastParty struct {
	db  *reform.DB
	dbx *sqlx.DB
}

func NewLastParty(db *reform.DB, dbx *sqlx.DB) *LastParty {
	return &LastParty{db, dbx}
}

func (x *LastParty) Party(_ struct{}, r *data.Party) error {
	return data.GetLastPartyWithProductsInfo(x.db, data.ProductsFilter{}, r)
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

func (x LastParty) SelectAll(checked [1]bool, _ *struct{}) (err error) {
	_, err = x.db.Exec(`
UPDATE product SET production = ? WHERE party_id = (SELECT last_party.party_id FROM last_party)`, checked[0])
	return
}

func (x LastParty) Export(_ struct{}, _ *struct{}) error {
	return data.ExportLastParty(x.db)
}

func (x *LastParty) Import(_ struct{}, r *data.Party) (err error) {
	return data.ImportLastParty(x.db)
}

func (x *LastParty) GetCheckBlocks(_ struct{}, r *GetCheckBlocksArg) error {
	return data.GetBlocksChecked(x.dbx, &r.Check)
}

func (x *LastParty) SetBlockChecked(r [2]int, a *int64) error {
	if err := data.SetBlockChecked(x.dbx, r[0], r[1] != 0); err != nil {
		return err
	}
	b := false
	if err := data.GetBlockChecked(x.dbx, r[0], &b); err != nil {
		return err
	}
	if b {
		*a = 1
	}
	return nil
}

func (x *LastParty) Pdf(_ struct{}, _ *struct{}) error {
	return pdf.Run(x.db)
}

func (x *LastParty) CalculateFonMinus20(_ struct{}, party *data.Party) error {
	if err := data.CalculateFonMinus20(x.db, x.dbx); err != nil {
		return err
	}
	return data.GetLastPartyWithProductsInfo(x.db, data.ProductsFilter{}, party)
}

func (x *LastParty) CalculateSensMinus20(k [1]float64, party *data.Party) error {
	if err := data.CalculateSensMinus20(x.db, x.dbx, k[0]); err != nil {
		return err
	}
	return data.GetLastPartyWithProductsInfo(x.db, data.ProductsFilter{}, party)
}

func (x *LastParty) CalculateSensPlus50(k [1]float64, party *data.Party) error {
	if err := data.CalculateSensPlus50(x.db, x.dbx, k[0]); err != nil {
		return err
	}
	return data.GetLastPartyWithProductsInfo(x.db, data.ProductsFilter{}, party)
}
