package api

import (
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/data/chipmem"
	"strings"
)

type LastPartySvc struct {
	RunWork func(name string, work func() error)
}

func (_ *LastPartySvc) Party(_ struct{}, r *Party1) error {
	*r = LastParty1()
	return nil
}

func (_ *LastPartySvc) GetValues(_ struct{}, p *Party3) error {
	*p = newParty3(data.LastParty())
	return nil
}

func (_ *LastPartySvc) SetProductType(ptName [1]string, _ *struct{}) error {
	var pt data.ProductType
	if err := data.DB.FindByPrimaryKeyTo(&pt, ptName[0]); err != nil {
		return err
	}
	p := data.LastParty()

	p.ProductTypeName = pt.ProductTypeName
	p.MinFon = pt.MinFon
	p.MaxFon = pt.MaxFon
	p.MaxDFon = pt.MaxDFon
	p.MinKSens20 = pt.MinKSens20
	p.MaxKSens20 = pt.MaxKSens20
	p.MinKSens50 = pt.MinKSens50
	p.MaxKSens50 = pt.MaxKSens50
	p.MinDTemp = pt.MinDTemp
	p.MaxDTemp = pt.MaxDTemp
	p.MaxDNotMeasured = pt.MaxDNotMeasured
	p.PointsMethod = pt.PointsMethod

	p.MaxD1 = pt.MaxD1
	p.MaxD2 = pt.MaxD2
	p.MaxD3 = pt.MaxD3
	return data.DB.Save(&p)
}

func (_ *LastPartySvc) SetValues(r struct{ P Party3 }, _ *struct{}) error {
	x := r.P

	p := data.LastParty()
	if err := x.SetupDataParty(&p); err != nil {
		return err
	}

	p.MinFon = stringToNullFloat(x.MinFon)
	p.MaxFon = stringToNullFloat(x.MaxFon)
	p.MaxDFon = stringToNullFloat(x.MaxDFon)
	p.MinKSens20 = stringToNullFloat(x.MinKSens20)
	p.MaxKSens20 = stringToNullFloat(x.MaxKSens20)
	p.MinKSens50 = stringToNullFloat(x.MinKSens50)
	p.MaxKSens50 = stringToNullFloat(x.MaxKSens50)
	p.MinDTemp = stringToNullFloat(x.MinDTemp)
	p.MaxDTemp = stringToNullFloat(x.MaxDTemp)
	p.MaxDNotMeasured = stringToNullFloat(x.MaxDNotMeasured)
	p.PointsMethod = x.PointsMethod
	p.MaxD1 = stringToNullFloat(x.MaxD1)
	p.MaxD2 = stringToNullFloat(x.MaxD2)
	p.MaxD3 = stringToNullFloat(x.MaxD3)

	if err := data.DB.Save(&p); err != nil {
		return err
	}

	var pt data.ProductType

	if err := data.DB.FindByPrimaryKeyTo(&pt, x.ProductTypeName); err != nil {
		return err
	}

	pt.MinFon = p.MinFon
	pt.MaxFon = p.MaxFon
	pt.MaxDFon = p.MaxDFon
	pt.MinKSens20 = p.MinKSens20
	pt.MaxKSens20 = p.MaxKSens20
	pt.MinKSens50 = p.MinKSens50
	pt.MaxKSens50 = p.MaxKSens50
	pt.MinDTemp = p.MinDTemp
	pt.MaxDTemp = p.MaxDTemp
	pt.MaxDNotMeasured = p.MaxDNotMeasured
	pt.PointsMethod = p.PointsMethod
	pt.MaxD1 = p.MaxD1
	pt.MaxD2 = p.MaxD2
	pt.MaxD3 = p.MaxD3

	if err := data.DB.Save(&pt); err != nil {
		return err
	}

	return nil
}

func (x *LastPartySvc) PartyID(_ struct{}, r *int64) error {
	*r = data.LastPartyID()
	return nil
}

func (x *LastPartySvc) SelectOnlyOkProductsProduction(_ struct{}, r *Party1) error {
	data.SetOnlyOkProductsProduction()
	*r = LastParty1()
	return nil
}

func (x *LastPartySvc) SetProductSerialAtPlace(p [2]int, r *int64) (err error) {
	*r, err = data.UpdateProductAtPlace(p[0], func(product *data.Product) {
		product.Serial.Int64 = int64(p[1])
		product.Serial.Valid = true
	})
	return
}

func (x LastPartySvc) ProductAtPlace(place [1]int, r *data.ProductInfo) error {
	partyID := data.LastPartyID()
	return data.DB.SelectOneTo(r, "WHERE party_id = ? AND place = ?", partyID, place)
}

func (x LastPartySvc) ToggleProductProductionAtPlace(place [1]int, r *int64) (err error) {
	*r, err = data.UpdateProductAtPlace(place[0], func(p *data.Product) {
		p.Production = !p.Production
	})
	return
}

func (x LastPartySvc) SetProductNoteAtPlace(p struct {
	Place int
	Note  string
}, r *int64) (err error) {
	*r, err = data.UpdateProductAtPlace(p.Place, func(product *data.Product) {
		product.Note.String = strings.TrimSpace(p.Note)
		product.Note.Valid = len(product.Note.String) > 0
	})
	return
}

func (x LastPartySvc) SetProductTypeAtPlacesRange(p struct {
	Place1, Place2 int
	ProductType    string
}, _ *struct{}) (err error) {
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
}, _ *struct{}) error {
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
	_, err = data.DBx.Exec(`
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

func (x *LastPartySvc) CalculateMinus20(k [1]float64, _ *struct{}) error {
	x.RunWork("расчёт токов -20", func() error {
		return chipmem.CalculateMinus20(k[0])
	})
	return nil
}

func (x *LastPartySvc) CalculateSensPlus50(k [1]float64, party *Party1) error {
	x.RunWork("расчёт тока чувствительности +50", func() error {
		return chipmem.CalculateSensPlus50(k[0])
	})
	return nil
}
