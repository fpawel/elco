package data

import (
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/sirupsen/logrus"
	"gopkg.in/reform.v1"
)

//go:generate go run github.com/fpawel/elco/cmd/utils/sqlstr/...

func EnsureProductTypeName(productTypeName string) error {
	_, err := DB.Exec(`
INSERT OR IGNORE INTO product_type 
  (product_type_name, gas_name, units_name, scale, noble_metal_content, lifetime_months)
VALUES (?, 'CO', 'мг/м3', 200, 0.1626, 18)`, productTypeName)
	return merry.Wrap(err)
}

func GetLastParty() (party Party) {
	err := DB.SelectOneTo(&party, `ORDER BY created_at DESC LIMIT 1;`)
	if err == reform.ErrNoRows {
		partyID := CreateNewParty()
		err = DB.FindByPrimaryKeyTo(&party, partyID)
	}
	if err != nil {
		panic(err)
	}
	party.Last = true
	return
}

func SetOnlyOkProductsProduction() {
	DBx.MustExec(`
UPDATE product 
	SET production = (SELECT ok FROM product_info WHERE product_info.product_id = product.product_id)
	WHERE party_id = (SELECT last_party.party_id FROM last_party)`)

}

func CreateNewParty() int64 {
	r, err := DB.Exec(`INSERT INTO party DEFAULT VALUES`)
	if err != nil {
		panic(err)
	}
	partyID, err := r.LastInsertId()
	if err != nil {
		panic(err)
	}
	if r, err = DB.Exec(`INSERT INTO product(party_id, serial, place) VALUES (?, 1, 0)`, partyID); err != nil {
		panic(err)
	}
	logrus.Warnf("new party created: %d", partyID)
	return partyID
}

func GetLastPartyID() (partyID int64) {
	row := DB.QueryRow(`SELECT party_id FROM party ORDER BY created_at DESC LIMIT 1`)

	if err := row.Scan(&partyID); err == sql.ErrNoRows {
		return CreateNewParty()
	}
	return partyID
}

type ProductsFilter struct {
	WithSerials, WithProduction bool
}

func GetLastPartyWithProductsInfo(f ProductsFilter) (party Party) {

	party = GetLastParty()

	tail := "WHERE party_id IN (SELECT party_id FROM last_party)"
	if f.WithSerials {
		tail += " AND (serial NOTNULL)"
	}
	if f.WithProduction {
		tail += " AND production"
	}
	tail += " ORDER BY place"
	xs, err := DB.SelectAllFrom(ProductInfoTable, tail)
	if err != nil {
		panic(err)
	}
	for _, x := range xs {
		p := x.(*ProductInfo)
		party.Products = append(party.Products, *p)
	}
	return party
}

func GetLastPartyProducts(f ProductsFilter) []Product {
	tail := "WHERE party_id IN (SELECT party_id FROM last_party)"
	if f.WithSerials {
		tail += " AND (serial NOTNULL)"
	}
	if f.WithProduction {
		tail += " AND production"
	}
	tail += " ORDER BY place"
	xs, err := DB.SelectAllFrom(ProductTable, tail)
	if err != nil {
		panic(err)
	}
	return structToProductSlice(xs)
}

func structToProductSlice(xs []reform.Struct) (products []Product) {
	for _, x := range xs {
		p := x.(*Product)
		products = append(products, *p)
	}
	return
}

func GetProductsInfoWithPartyID(partyID int64) []ProductInfo {
	xs, err := DB.SelectAllFrom(ProductInfoTable, "WHERE party_id = ? ORDER BY place", partyID)
	if err != nil {
		panic(err)
	}
	var productsInfo []ProductInfo
	for _, x := range xs {
		productsInfo = append(productsInfo, *x.(*ProductInfo))
	}
	return productsInfo
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

func Gases() []Gas {
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

func GetLastPartyProductAtPlace(place int, product *Product) error {
	return DB.SelectOneTo(product, "WHERE party_id = (SELECT party_id FROM last_party) AND place = ?", place)
}

func GetProductAtPlace(place int, product *Product) (err error) {
	err = DB.SelectOneTo(product, "WHERE party_id = ? AND place = ?", GetLastPartyID(), place)
	return
}

func UpdateProductAtPlace(place int, f func(p *Product) error) (int64, error) {
	partyID := GetLastPartyID()

	var p Product
	if err := DB.SelectOneTo(&p, "WHERE party_id = ? AND place = ?", partyID, place); err != nil && err != reform.ErrNoRows {
		return 0, err
	}
	if err := f(&p); err != nil {
		return 0, err
	}
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

func GetProductByProductID(productID int64) Product {
	var p Product
	if err := DB.SelectOneTo(&p, `WHERE product_id = ?`, productID); err != nil {
		panic(err)
	}
	return p
}

func GetProductInfoByProductID(productID int64) ProductInfo {
	var p ProductInfo
	if err := DB.SelectOneTo(&p, `WHERE product_id = ?`, productID); err != nil {
		panic(err)
	}
	return p
}
