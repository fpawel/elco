package data

import (
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"gopkg.in/reform.v1"
	"time"
)

//go:generate go run github.com/fpawel/elco/cmd/utils/sqlstr/...

func EnsureProductTypeName(productTypeName string) error {
	_, err := DB.Exec(`
INSERT OR IGNORE INTO product_type 
  (product_type_name, gas_name, units_name, scale, noble_metal_content, lifetime_months)
VALUES (?, 'CO', 'мг/м3', 200, 0.1626, 18)`, productTypeName)
	return merry.Wrap(err)
}

func SetOnlyOkProductsProduction() {
	DBx.MustExec(`
UPDATE product 
	SET production = (SELECT ok FROM product_info WHERE product_info.product_id = product.product_id)
	WHERE party_id = (SELECT last_party.party_id FROM last_party)`)

}

func CreateNewParty() int64 {
	party := Party{
		CreatedAt:       time.Now(),
		ProductTypeName: "035",
		PointsMethod:    3,
		Concentration1:  0,
		Concentration2:  50,
		Concentration3:  100,
		MinFon:          sql.NullFloat64{-1, true},
		MaxFon:          sql.NullFloat64{2, true},
		MaxDFon:         sql.NullFloat64{3, true},
		MinKSens20:      sql.NullFloat64{0.08, true},
		MaxKSens20:      sql.NullFloat64{0.335, true},
		MinKSens50:      sql.NullFloat64{110, true},
		MaxKSens50:      sql.NullFloat64{150, true},
		MaxDTemp:        sql.NullFloat64{3, true},
	}
	if err := DB.Save(&party); err != nil {
		panic(err)
	}
	return party.PartyID
}

func GetLastPartyID() (partyID int64) {
	row := DB.QueryRow(`SELECT party_id FROM party ORDER BY created_at DESC LIMIT 1`)

	if err := row.Scan(&partyID); err == sql.ErrNoRows {
		return CreateNewParty()
	}
	return partyID
}

type ProductsFilter int

const (
	WithSerials ProductsFilter = iota
	WithProduction
)

func productsFilterQuery(f ...ProductsFilter) (tail string) {
	if productsFilter(WithSerials, f...) {
		tail += " AND (serial NOTNULL)"
	}
	if productsFilter(WithProduction, f...) {
		tail += " AND production"
	}
	tail += " ORDER BY place"
	return
}

func productsFilter(y ProductsFilter, f ...ProductsFilter) bool {
	for _, x := range f {
		if x == y {
			return true
		}
	}
	return false
}

type PartyProducts bool

const (
	WithoutProducts PartyProducts = false
	WithProducts    PartyProducts = true
)

func GetParty(partyID int64, withProducts PartyProducts, f ...ProductsFilter) (party Party, err error) {
	if err = DB.FindByPrimaryKeyTo(&party, partyID); err != nil {
		return
	}
	party.Last = party.PartyID == GetLastPartyID()
	if withProducts {
		tail := "WHERE party_id = ?" + productsFilterQuery(f...)
		xs, err := DB.SelectAllFrom(ProductInfoTable, tail, partyID)
		if err != nil {
			panic(err)
		}
		for _, x := range xs {
			p := x.(*ProductInfo)
			party.Products = append(party.Products, *p)
		}
	}
	return
}

func GetLastParty(withProducts PartyProducts, f ...ProductsFilter) Party {
	partyID := GetLastPartyID()
	party, err := GetParty(partyID, withProducts, f...)
	if err != nil {
		panic(err)
	}
	return party
}

func GetLastPartyProducts(f ...ProductsFilter) []Product {
	tail := "WHERE party_id IN (SELECT party_id FROM last_party)" +
		productsFilterQuery(f...)

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