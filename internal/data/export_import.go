package data

import (
	"encoding/json"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/elco"
	"gopkg.in/reform.v1"
	"io/ioutil"
	"time"
)

func ExportLastParty(db *reform.DB) error {

	party, err := GetLastParty(db)
	if err != nil {
		return err
	}
	products, err := GetLastPartyProducts(db, ProductsFilter{})
	if err != nil {
		return err
	}
	oldParty := party.OldParty(products)
	b, err := json.MarshalIndent(&oldParty, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(importFileName(), b, 0666)

}

func ImportLastParty(db *reform.DB) error {

	b, err := ioutil.ReadFile(importFileName())
	if err != nil {
		return err
	}
	var oldParty OldParty
	if err := json.Unmarshal(b, &oldParty); err != nil {
		return err
	}
	party, products := oldParty.Party()

	if err := EnsureProductTypeName(db, party.ProductTypeName); err != nil {
		return err
	}
	party.CreatedAt = time.Now().Add(-3 * time.Hour)
	if err := db.Save(&party); err != nil {
		return err
	}
	for _, p := range products {
		p.PartyID = party.PartyID
		if p.ProductTypeName.Valid {
			if err := EnsureProductTypeName(db, p.ProductTypeName.String); err != nil {
				return err
			}
		}
		if err := db.Save(&p); err != nil {
			return merry.Appendf(err, "product: serial: %v place: %d",
				p.Serial, p.Place)
		}
	}
	return nil
}

func importFileName() string {
	return elco.AppName.DataFileName("export-party.json")
}
