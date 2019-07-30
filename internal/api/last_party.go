package api

import (
	"database/sql"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/data/old"
	"strings"
)

type LastPartySvc struct {
}

func (x *LastPartySvc) Party(_ struct{}, r *data.Party) error {
	*r = data.GetLastParty(data.WithAllProducts)
	return nil
}

type LastPartyValues struct {
	ProductTypeName string
	Concentration1  float64
	Concentration2  float64
	Concentration3  float64
	Note            sql.NullString
	MinFon          sql.NullFloat64
	MaxFon          sql.NullFloat64
	MaxDFon         sql.NullFloat64
	MinKSens20      sql.NullFloat64
	MaxKSens20      sql.NullFloat64
	MinKSens50      sql.NullFloat64
	MaxKSens50      sql.NullFloat64
	MinDTemp        sql.NullFloat64
	MaxDTemp        sql.NullFloat64
	MaxDNotMeasured sql.NullFloat64
	PointsMethod    int64
}

func (x *LastPartySvc) SetValues(r struct{ Values LastPartyValues }, _ *struct{}) error {
	p := data.GetLastParty(data.WithoutProducts)
	y := r.Values
	p.ProductTypeName = y.ProductTypeName
	p.Concentration1 = y.Concentration1
	p.Concentration2 = y.Concentration2
	p.Concentration3 = y.Concentration3
	p.Note = y.Note
	p.MinFon = y.MinFon
	p.MaxFon = y.MaxFon
	p.MaxDFon = y.MaxDFon
	p.MinKSens20 = y.MinKSens20
	p.MaxKSens20 = y.MaxKSens20
	p.MinKSens50 = y.MinKSens50
	p.MaxKSens50 = y.MaxKSens50
	p.MinDTemp = y.MinDTemp
	p.MaxDTemp = y.MaxDTemp
	p.MaxDNotMeasured = y.MaxDNotMeasured
	p.PointsMethod = y.PointsMethod
	return data.DB.Save(&p)
}

func (x *LastPartySvc) PartyID(_ struct{}, r *int64) error {
	*r = data.GetLastPartyID()
	return nil
}

func (x *LastPartySvc) SelectOnlyOkProductsProduction(_ struct{}, r *data.Party) error {
	data.SetOnlyOkProductsProduction()
	*r = data.GetLastParty(data.WithAllProducts)
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

func (x LastPartySvc) SetProductTypeAtPlacesRange(p struct {
	Place1, Place2 int
	ProductType    string
}, r *int64) (err error) {
	p.ProductType = strings.TrimSpace(p.ProductType)
	var v interface{}
	if p.ProductType != "" {
		v = p.ProductType
	}
	data.DBx.MustExec(`
UPDATE product 
SET product_type_name=? 
WHERE party_id=(SELECT party_id FROM last_party) 
  AND place BETWEEN ? AND ?`, v, p.Place1, p.Place2)
	return nil
}

func (x LastPartySvc) SetPointsMethodInPlacesRange(p struct {
	Place1, Place2 int
	Value          int
}, r *int64) error {
	var v interface{}
	if p.Value == 2 || p.Value == 3 {
		v = p.Value
	}
	data.DBx.MustExec(`
UPDATE product 
SET points_method=? 
WHERE party_id=(SELECT party_id FROM last_party) 
  AND place BETWEEN ? AND ?`, v, p.Place1, p.Place2)
	return nil
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
	*party = data.GetLastParty(data.WithAllProducts)
	return nil
}

func (x *LastPartySvc) CalculateSensMinus20(k [1]float64, party *data.Party) error {
	if err := data.CalculateSensMinus20(k[0]); err != nil {
		return err
	}
	*party = data.GetLastParty(data.WithAllProducts)
	return nil
}

func (x *LastPartySvc) CalculateSensPlus50(k [1]float64, party *data.Party) error {
	if err := data.CalculateSensPlus50(k[0]); err != nil {
		return err
	}
	*party = data.GetLastParty(data.WithAllProducts)
	return nil
}
