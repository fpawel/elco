package crud

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/settings"
	"github.com/fpawel/goutils/dbutils"
	"github.com/pkg/errors"
	"gopkg.in/reform.v1"
	"io/ioutil"
	"strconv"
	"strings"
)

type LastParty struct {
	dbContext
}

func (x LastParty) Party() data.Party {
	x.mu.Lock()
	defer x.mu.Unlock()
	return data.LastParty(x.dbr)
}

func (x LastParty) ProductsWithProduction() []data.Product {
	x.mu.Lock()
	defer x.mu.Unlock()
	return data.GetLastPartyProductsWithProduction(x.dbr)
}

func (x LastParty) ProductsWithSerials() []data.Product {
	x.mu.Lock()
	defer x.mu.Unlock()
	return data.GetLastPartyProductsWithSerials(x.dbr)
}

func (x LastParty) ProductAtPlace(place int) (product data.ProductInfo, err error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	party := data.LastParty(x.dbr)
	err = x.dbr.SelectOneTo(&product, "WHERE party_id = ? AND place = ?", party.PartyID, place)
	return
}

func (x LastParty) updateProductAtPlace(place int, f func(p *data.Product) error) (int64, error) {
	partyID := x.partyID()
	var p data.Product
	if err := x.dbr.SelectOneTo(&p, "WHERE party_id = ? AND place = ?", partyID, place); err != nil && err != reform.ErrNoRows {
		return 0, err
	}
	if err := f(&p); err != nil {
		return 0, err
	}
	p.PartyID = partyID
	p.Place = place
	if err := x.dbr.Save(&p); err != nil {
		return 0, err
	}
	return p.ProductID, nil
}

func (x LastParty) SetProductSerialAtPlace(place, serial int) (int64, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.updateProductAtPlace(place, func(p *data.Product) error {
		p.Serial.Int64 = int64(serial)
		p.Serial.Valid = true
		return nil
	})
}

func (x LastParty) SetProductNoteAtPlace(place int, note string) (int64, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.updateProductAtPlace(place, func(p *data.Product) error {
		p.Note.String = strings.TrimSpace(note)
		p.Note.Valid = len(p.Note.String) > 0
		return nil
	})
}

func (x LastParty) SetPointsMethodAtPlace(place int, pointsMethod int64, valid bool) (int64, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.updateProductAtPlace(place, func(p *data.Product) error {
		p.PointsMethod.Int64 = pointsMethod
		p.PointsMethod.Valid = valid
		return nil
	})
}

func (x LastParty) SetProductTypeAtPlace(place int, productType string) (int64, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.updateProductAtPlace(place, func(p *data.Product) error {
		p.ProductTypeName.String = strings.TrimSpace(productType)
		p.ProductTypeName.Valid = len(p.ProductTypeName.String) > 0
		return nil
	})
}

func (x LastParty) ToggleProductProductionAtPlace(place int) (int64, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.updateProductAtPlace(place, func(p *data.Product) error {
		p.Production = !p.Production
		return nil
	})
}

func (x LastParty) DeleteProductAtPlace(place int) (err error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	partyID := x.partyID()
	_, err = x.dbx.Exec(`DELETE FROM product WHERE party_id = ? AND place = ?`, partyID, place)
	return
}

func (x LastParty) SetConfigValue(property, value string) (err error) {
	x.mu.Lock()
	defer x.mu.Unlock()

	party := data.LastParty(x.dbr)

	parseFloat := func() (float64, error) {
		return strconv.ParseFloat(strings.Replace(value, ",", ".", -1), 64)
	}

	switch property {

	case "ProductType":
		party.ProductTypeName = value
		return x.dbr.Save(&party)

	case "Gas1":
		party.Concentration1, err = parseFloat()
		if err == nil {
			err = x.dbr.Save(&party)
		}
		return

	case "Gas2":
		party.Concentration2, err = parseFloat()
		if err == nil {
			err = x.dbr.Save(&party)
		}
		return

	case "Gas3":
		party.Concentration3, err = parseFloat()
		if err == nil {
			err = x.dbr.Save(&party)
		}
		return

	case "Note":
		party.Note.String = strings.TrimSpace(value)
		party.Note.Valid = len(party.Note.String) > 0
		err = x.dbr.Save(&party)
		return

	case "PointsMethod":
		party.PointsMethod, err = strconv.ParseInt(value, 10, 8)
		if err != nil {
			return err
		}
		err = x.dbr.Save(&party)
		return

	default:
		var v sql.NullFloat64
		fs := map[string]func(){
			"MinFon": func() {
				party.MinFon = v
			},
			"MaxFon": func() {
				party.MaxFon = v
			},
			"MaxDFon": func() {
				party.MaxDFon = v
			},
			"MinKSens20": func() {
				party.MinKSens20 = v
			},
			"MaxKSens20": func() {
				party.MaxKSens20 = v
			},
			"MinKSens50": func() {
				party.MinKSens50 = v
			},
			"MaxKSens50": func() {
				party.MaxKSens50 = v
			},
			"MinDTemp": func() {
				party.MinDTemp = v
			},
			"MaxDTemp": func() {
				party.MaxDTemp = v
			},
			"MaxDNotMeasured": func() {
				party.MaxDNotMeasured = v
			},
		}
		if f, ok := fs[property]; ok {
			if len(strings.TrimSpace(value)) > 0 {
				if v.Float64, err = parseFloat(); err != nil {
					return err
				}
				v.Valid = true
			}
			f()
			return x.dbr.Save(&party)
		}
	}
	return errors.Errorf("%q: wrong party property")
}

func (x LastParty) ConfigProperties() []settings.ConfigProperty {
	party := x.Party()
	productTypesNames := x.ListProductTypesNames()

	f := func(v sql.NullFloat64) string {
		if v.Valid {
			return fmt.Sprintf("%v", v.Float64)
		}
		return ""
	}

	return []settings.ConfigProperty{
		{
			Hint:      "Исполнение",
			Name:      "ProductType",
			ValueType: settings.VtString,
			Value:     party.ProductTypeName,
			List:      productTypesNames,
		},
		{
			Hint:      "ПГС1",
			Name:      "Gas1",
			ValueType: settings.VtFloat,
			Value:     fmt.Sprintf("%v", party.Concentration1),
		},
		{
			Hint:      "ПГС2",
			Name:      "Gas2",
			ValueType: settings.VtFloat,
			Value:     fmt.Sprintf("%v", party.Concentration2),
		},
		{
			Hint:      "ПГС3",
			Name:      "Gas3",
			ValueType: settings.VtFloat,
			Value:     fmt.Sprintf("%v", party.Concentration3),
		},
		{
			Hint:      "Примечание",
			Name:      "Note",
			ValueType: settings.VtString,
			Value:     fmt.Sprintf("%v", party.Note.String),
		},

		{
			Hint:      "Кол-во точек для расчёта",
			Name:      "PointsMethod",
			ValueType: settings.VtString,
			Value:     strconv.Itoa(int(party.PointsMethod)),
			List:      []string{"2", "3"},
		},

		{
			Hint:      "Фон.мин, мкА",
			Name:      "MinFon",
			ValueType: settings.VtNullFloat,
			Value:     f(party.MinFon),
		},
		{
			Hint:      "Фон.мax, мкА",
			Name:      "MaxFon",
			ValueType: settings.VtNullFloat,
			Value:     f(party.MaxFon),
		},
		{
			Hint:      "D.фон.мax, мкА",
			Name:      "MaxDFon",
			ValueType: settings.VtNullFloat,
			Value:     f(party.MaxDFon),
		},
		{
			Hint:      "Кч20.мин, мкА/мг/м3",
			Name:      "MinKSens20",
			ValueType: settings.VtNullFloat,
			Value:     f(party.MinKSens20),
		},
		{
			Hint:      "Кч20.макс, мкА/мг/м3",
			Name:      "MaxKSens20",
			ValueType: settings.VtNullFloat,
			Value:     f(party.MaxKSens20),
		},
		{
			Hint:      "Кч50.мин, мкА/мг/м3",
			Name:      "MinKSens50",
			ValueType: settings.VtNullFloat,
			Value:     f(party.MinKSens50),
		},
		{
			Hint:      "Кч50.макс, мкА/мг/м3",
			Name:      "MaxKSens50",
			ValueType: settings.VtNullFloat,
			Value:     f(party.MaxKSens50),
		},
		{
			Hint:      "Dt.мин, мкА",
			Name:      "MinDTemp",
			ValueType: settings.VtNullFloat,
			Value:     f(party.MinDTemp),
		},
		{
			Hint:      "Dt.мин, мкА",
			Name:      "MaxDTemp",
			ValueType: settings.VtNullFloat,
			Value:     f(party.MaxDTemp),
		},
		{
			Hint:      "Dn.макс, мкА",
			Name:      "MaxDNotMeasured",
			ValueType: settings.VtNullFloat,
			Value:     f(party.MaxDNotMeasured),
		},
	}
}

func (x LastParty) partyID() (lastPartyID int64) {
	dbutils.MustGet(x.dbx, &lastPartyID, `SELECT party_id FROM last_party`)
	return
}

func (x LastParty) ExportToFile() error {
	x.mu.Lock()
	defer x.mu.Unlock()

	party := data.LastParty(x.dbr)
	products := data.GetLastPartyProducts(x.dbr)
	oldParty := party.OldParty(products)
	b, err := json.MarshalIndent(&oldParty, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(importFileName(), b, 0666)

}
