package api

import (
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/data/old"
	"strings"
)

type LastPartySvc struct {
}

func (x *LastPartySvc) Party(_ struct{}, r *data.Party) error {
	*r = data.GetLastPartyWithProductsInfo(data.ProductsFilter{})
	return nil
}

func (x *LastPartySvc) SelectOnlyOkProductsProduction(_ struct{}, r *data.Party) error {
	data.SetOnlyOkProductsProduction()
	*r = data.GetLastPartyWithProductsInfo(data.ProductsFilter{})
	return nil
}

func (x *LastPartySvc) SetProductSerialAtPlace(p [2]int, r *int64) (err error) {
	*r, err = data.UpdateProductAtPlace(p[0], func(product *data.Product) error {
		product.Serial.Int64 = int64(p[1])
		product.Serial.Valid = true
		return nil
	})
	return
}

func (x LastPartySvc) ProductAtPlace(place [1]int, r *data.ProductInfo) error {
	partyID := data.GetLastPartyID()
	return data.DB.SelectOneTo(r, "WHERE party_id = ? AND place = ?", partyID, place)
}

func (x LastPartySvc) ToggleProductProductionAtPlace(place [1]int, r *int64) (err error) {
	*r, err = data.UpdateProductAtPlace(place[0], func(p *data.Product) error {
		p.Production = !p.Production
		return nil
	})
	return
}

func (x LastPartySvc) SetProductNoteAtPlace(p struct {
	Place int
	Note  string
}, r *int64) (err error) {
	*r, err = data.UpdateProductAtPlace(p.Place, func(product *data.Product) error {
		product.Note.String = strings.TrimSpace(p.Note)
		product.Note.Valid = len(product.Note.String) > 0
		return nil
	})
	return
}

func (x LastPartySvc) SetProductTypeAtPlace(p struct {
	Place       int
	ProductType string
}, r *int64) (err error) {
	*r, err = data.UpdateProductAtPlace(p.Place, func(product *data.Product) error {
		product.ProductTypeName.String = strings.TrimSpace(p.ProductType)
		product.ProductTypeName.Valid = len(product.ProductTypeName.String) > 0
		return nil
	})
	return
}

func (x LastPartySvc) SetPointsMethodAtPlace(p struct {
	Place        int
	PointsMethod int64
	Valid        bool
}, r *int64) (err error) {
	*r, err = data.UpdateProductAtPlace(p.Place, func(product *data.Product) error {
		product.PointsMethod.Int64 = p.PointsMethod
		product.PointsMethod.Valid = p.Valid
		return nil
	})
	return
}

func (x LastPartySvc) DeleteProductAtPlace(place [1]int, _ *struct{}) (err error) {
	_, err = data.DB.Exec(`
DELETE FROM product 
WHERE party_id IN (
  SELECT party.party_id 
  FROM party 
  ORDER BY created_at DESC 
  LIMIT 1) AND place = ?`, place[0])
	return
}

func (x LastPartySvc) SelectAll(checked [1]bool, _ *struct{}) (err error) {
	_, err = data.DB.Exec(`
UPDATE product SET production = ? WHERE party_id = (SELECT last_party.party_id FROM last_party)`, checked[0])
	return
}

func (x LastPartySvc) Export(_ struct{}, _ *struct{}) error {
	return old.ExportLastParty()
}

func (x *LastPartySvc) Import(_ struct{}, r *data.Party) (err error) {
	return old.ImportLastParty()
}

func (x *LastPartySvc) GetCheckBlocks(_ struct{}, r *GetCheckBlocksArg) error {
	return data.GetBlocksChecked(&r.Check)
}

func (x *LastPartySvc) SetBlockChecked(r [2]int, a *int64) error {
	data.SetBlockChecked(r[0], r[1] != 0)
	b := data.GetBlockChecked(r[0])
	if b {
		*a = 1
	}
	return nil
}

func (x *LastPartySvc) CalculateFonMinus20(_ struct{}, party *data.Party) error {
	if err := data.CalculateFonMinus20(); err != nil {
		return err
	}
	*party = data.GetLastPartyWithProductsInfo(data.ProductsFilter{})
	return nil
}

func (x *LastPartySvc) CalculateSensMinus20(k [1]float64, party *data.Party) error {
	if err := data.CalculateSensMinus20(k[0]); err != nil {
		return err
	}
	*party = data.GetLastPartyWithProductsInfo(data.ProductsFilter{})
	return nil
}

func (x *LastPartySvc) CalculateSensPlus50(k [1]float64, party *data.Party) error {
	if err := data.CalculateSensPlus50(k[0]); err != nil {
		return err
	}
	*party = data.GetLastPartyWithProductsInfo(data.ProductsFilter{})
	return nil
}
