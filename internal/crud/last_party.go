package crud

import (
	"fmt"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/settings"
	"github.com/fpawel/goutils/dbutils"
	"github.com/pkg/errors"
	"gopkg.in/reform.v1"
	"strconv"
	"strings"
)

type LastParty struct {
	dbContext
}

func (x LastParty) Party() data.Party {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.party()
}

func (x LastParty) ProductAtPlace(place int) (product data.ProductInfo, err error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	party := x.party()
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
	_, err = x.dbr.DeleteFrom(data.ProductTable, "WHERE party_id = ? AND place = ?", partyID, place)
	return
}

func (x LastParty) SetConfigValue(property, value string) (err error) {
	x.mu.Lock()
	defer x.mu.Unlock()

	party := x.party()

	switch property {
	case "ProductType":
		party.ProductTypeName = value
		return x.dbr.Save(&party)
	case "Gas1":
		party.Concentration1, err = strconv.ParseFloat(value, 64)
		if err == nil {
			err = x.dbr.Save(&party)
		}
		return
	case "Gas2":
		party.Concentration2, err = strconv.ParseFloat(value, 64)
		if err == nil {
			err = x.dbr.Save(&party)
		}
		return
	case "Gas3":
		party.Concentration3, err = strconv.ParseFloat(value, 64)
		if err == nil {
			err = x.dbr.Save(&party)
		}
		return
	}
	return errors.Errorf("%q: wrong party property")
}

func (x LastParty) ConfigProperties() []settings.ConfigProperty {
	party := x.party()
	productTypesNames := x.ListProductTypesNames()
	return []settings.ConfigProperty{
		{
			Hint:         "Исполнение",
			Name:         "ProductType",
			DefaultValue: productTypesNames[0],
			ValueType:    settings.VtString,
			Value:        party.ProductTypeName,
			List:         productTypesNames,
		},
		{
			Hint:         "ПГС1",
			Name:         "Gas1",
			DefaultValue: "0",
			ValueType:    settings.VtFloat,
			Value:        fmt.Sprintf("%v", party.Concentration1),
		},
		{
			Hint:         "ПГС2",
			Name:         "Gas2",
			DefaultValue: "100",
			ValueType:    settings.VtFloat,
			Value:        fmt.Sprintf("%v", party.Concentration2),
		},
		{
			Hint:         "ПГС3",
			Name:         "Gas3",
			DefaultValue: "200",
			ValueType:    settings.VtFloat,
			Value:        fmt.Sprintf("%v", party.Concentration3),
		},
	}
}

func (x LastParty) party() data.Party {
	var party data.Party
	if err := x.dbr.SelectOneTo(&party, `ORDER BY created_at DESC LIMIT 1;`); err != nil {
		panic(err)
	}
	party.Products = data.GetProductsByPartyID(x.dbr, party.PartyID)
	return party
}

func (x LastParty) partyID() (lastPartyID int64) {
	dbutils.MustGet(x.dbx, &lastPartyID, `SELECT party_id FROM last_party`)
	return
}
