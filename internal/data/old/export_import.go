package old

import (
	"encoding/json"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/data"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func ExportLastParty() error {
	party := data.GetLastParty()
	products := data.GetLastPartyProducts(data.ProductsFilter{})
	oldParty := NewOldParty(party, products)
	b, err := json.MarshalIndent(&oldParty, "", "    ")
	if err != nil {
		return err
	}
	importFileName := importFileName()
	return ioutil.WriteFile(importFileName, b, 0666)

}

func ImportLastParty() error {

	importFileName := importFileName()

	b, err := ioutil.ReadFile(importFileName)
	if err != nil {
		return err
	}
	var oldParty Party
	if err := json.Unmarshal(b, &oldParty); err != nil {
		return err
	}
	party, products := oldParty.Party()

	if err := data.EnsureProductTypeName(party.ProductTypeName); err != nil {
		return err
	}
	party.CreatedAt = time.Now().Add(-3 * time.Hour)
	if err := data.DB.Save(&party); err != nil {
		return err
	}
	for _, p := range products {
		p.PartyID = party.PartyID
		if p.ProductTypeName.Valid {
			if err := data.EnsureProductTypeName(p.ProductTypeName.String); err != nil {
				return err
			}
		}
		if err := data.DB.Save(&p); err != nil {
			return merry.Appendf(err, "product: serial: %v place: %d",
				p.Serial, p.Place)
		}
	}
	return nil
}

func importFileName() string {
	return filepath.Join(filepath.Dir(os.Args[0]), "export-party.json")
}
