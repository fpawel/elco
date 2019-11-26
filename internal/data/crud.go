package data

import (
	"database/sql"
	"fmt"
	"gopkg.in/reform.v1"
	"time"
)

//go:generate go run github.com/fpawel/gotools/cmd/sqlstr/...

func SetOnlyOkProductsProduction() {
	DBx.MustExec(`
UPDATE product 
	SET production = (SELECT ok FROM product_info WHERE product_info.product_id = product.product_id)
	WHERE party_id = (SELECT last_party.party_id FROM last_party)`)

}

func CreateNewParty() int64 {

	var pt ProductType

	if err := DB.FindByPrimaryKeyTo(&pt, "035"); err != nil {
		panic(err)
	}

	party := Party{
		CreatedAt:       time.Now(),
		ProductTypeName: "035",
		PointsMethod:    3,
		Concentration1:  0,
		Concentration2:  100,
		Concentration3:  200,
		MinFon:          pt.MinFon,
		MaxFon:          pt.MaxFon,
		MaxDFon:         pt.MaxDFon,
		MinKSens20:      pt.MinKSens20,
		MaxKSens20:      pt.MaxKSens20,
		MinKSens50:      pt.MinKSens50,
		MaxKSens50:      pt.MaxKSens50,
		MaxDTemp:        pt.MaxDTemp,
		MaxD1:           pt.MaxD1,
		MaxD2:           pt.MaxD2,
		MaxD3:           pt.MaxD3,
	}
	if err := DB.Save(&party); err != nil {
		panic(err)
	}
	return party.PartyID
}
func LastPartyID() (partyID int64) {
	row := DB.QueryRow(`SELECT party_id FROM party ORDER BY created_at DESC LIMIT 1`)
	if err := row.Scan(&partyID); err == sql.ErrNoRows {
		return CreateNewParty()
	}
	return partyID
}

//func ProductsAll(partyID int64) []Product {
//	return fetchProductsByPartyID(partyID, "WHERE party_id = ? ORDER BY place")
//}
func ProductsWithProduction(partyID int64) []Product {
	return fetchProductsByPartyID(partyID, "WHERE party_id = ? AND production ORDER BY place")
}
func ProductsInfoAll(partyID int64) []ProductInfo {
	return fetchProductsInfoByPartyID(partyID, "WHERE party_id = ? ORDER BY place")
}
func ProductsInfoWithProduction(partyID int64) []ProductInfo {
	return fetchProductsInfoByPartyID(partyID, "WHERE party_id = ? AND production ORDER BY place")
}
func GetParty(partyID int64) (party Party) {
	if err := DB.FindByPrimaryKeyTo(&party, partyID); err != nil {
		panic(err)
	}
	return
}
func LastParty() Party {
	return GetParty(LastPartyID())
}
func ProductTypeNames() []string {
	xs, err := DB.SelectAllFrom(ProductTypeTable, "ORDER BY product_type_name")
	if err != nil {
		panic(err)
	}
	var r []string
	for _, x := range xs {
		r = append(r, x.(*ProductType).ProductTypeName)
	}
	return r
}

func ListUnits() []Units {
	records, err := DB.SelectAllFrom(UnitsTable, "")
	if err != nil {
		panic(err)
	}
	var units []Units
	for _, r := range records {
		x := r.(*Units)
		units = append(units, *x)
	}
	return units
}

func ListGases() []Gas {
	records, err := DB.SelectAllFrom(GasTable, "")
	if err != nil {
		panic(err)
	}
	var gas []Gas
	for _, r := range records {
		x := r.(*Gas)
		gas = append(gas, *x)
	}
	return gas
}

func UpdateProductAtPlace(place int, f func(p *Product)) (int64, error) {
	partyID := LastPartyID()

	var p Product
	if err := DB.SelectOneTo(&p, "WHERE party_id = ? AND place = ?", partyID, place); err != nil && err != reform.ErrNoRows {
		return 0, err
	}
	f(&p)
	p.PartyID = partyID
	p.Place = place
	if err := DB.Save(&p); err != nil {
		return 0, err
	}
	return p.ProductID, nil
}

func GetLastPartyCheckedBlocks() (r []int) {
	err := DBx.Select(&r, `
WITH block AS (
  WITH RECURSIVE
    cnt(x) AS (
      SELECT 0
      UNION ALL
      SELECT x + 1
      FROM cnt
      LIMIT 12
      )
    SELECT x FROM cnt)

SELECT block.x AS block       
FROM block
WHERE EXISTS(
           SELECT *
           FROM product
           WHERE party_id = (SELECT party_id FROM last_party)
             AND production
             AND place / 8 = block.x) 
`)
	if err != nil {
		panic(err)
	}
	return
}

func GetBlocksChecked(r *[]bool) error {
	return DBx.Select(r, `
WITH block AS (
  WITH RECURSIVE
    cnt(x) AS (
      SELECT 0
      UNION ALL
      SELECT x + 1
      FROM cnt
      LIMIT 12
      )
    SELECT x FROM cnt)

SELECT EXISTS(
           SELECT *
           FROM product
           WHERE party_id = (SELECT party_id FROM last_party)
             AND production
             AND place / 8 = block.x) AS checked
FROM block`)
}

func GetBlockChecked(block int) (r bool) {
	if err := DBx.Get(&r, `
SELECT EXISTS( 
  SELECT * 
  FROM product 
  WHERE party_id = ( SELECT party_id FROM last_party) 
    AND production 
    AND place / 8 = ?)`, block); err != nil {
		panic(err)
	}
	return
}

func SetBlockChecked(block int, r bool) {
	DBx.MustExec(` 
  UPDATE product
  SET production = ?
  WHERE party_id = ( SELECT party_id FROM last_party) 
    AND place / 8 = ?`, r, block)
}

func SetProductValue(productID int64, field string, value float64) error {
	_, err := DBx.Exec(fmt.Sprintf(`UPDATE product SET %s = ? WHERE product_id = ?`, field), value, productID)
	return err
}

func fetchProductsInfoByPartyID(partyID int64, tail string) []ProductInfo {
	products := make([]ProductInfo, 0)
	xs, err := DB.SelectAllFrom(ProductInfoTable, tail, partyID)
	if err != nil {
		panic(err)
	}
	for _, x := range xs {
		products = append(products, *x.(*ProductInfo))
	}
	return products
}

func fetchProductsByPartyID(partyID int64, tail string) []Product {
	products := make([]Product, 0)
	xs, err := DB.SelectAllFrom(ProductTable, tail, partyID)
	if err != nil {
		panic(err)
	}
	for _, x := range xs {
		products = append(products, *x.(*Product))
	}
	return products
}
