package data

import (
	"database/sql"
	"github.com/ansel1/merry"
	"github.com/sirupsen/logrus"
	"gopkg.in/reform.v1"
)

//go:generate go run github.com/fpawel/elco/cmd/utils/sqlstr/...

func Init(db *reform.DB) error {
	_, err := db.Exec(SQLCreate)
	if err != nil {
		return merry.Wrap(err)
	}
	_, err = db.Exec(`
DELETE
FROM product
WHERE party_id NOT IN (SELECT party_id FROM last_party)  AND
  serial ISNULL  AND
  (product_type_name ISNULL OR LENGTH(trim(product_type_name)) = 0)  AND
  (note ISNULL OR LENGTH(trim(note)) = 0)  AND
  i_f_minus20 ISNULL  AND
  i_f_plus20 ISNULL  AND
  i_f_plus50 ISNULL  AND
  i_s_minus20 ISNULL  AND
  i_s_plus20 ISNULL  AND
  i_s_plus50 ISNULL  AND
  i13 ISNULL  AND
  i24 ISNULL  AND
  i35 ISNULL  AND
  i26 ISNULL  AND
  i17 ISNULL  AND
  not_measured ISNULL  AND
  (firmware ISNULL OR LENGTH(firmware) = 0)  AND
  old_product_id ISNULL  AND
  old_serial ISNULL;
DELETE
FROM party
WHERE NOT EXISTS(SELECT product_id FROM product WHERE party.party_id = product.party_id);
`)
	if err != nil {
		return merry.Wrap(err)
	}
	return nil
}

func EnsureProductTypeName(db *reform.DB, productTypeName string) error {
	_, err := db.Exec(`
INSERT OR IGNORE INTO product_type 
  (product_type_name, gas_name, units_name, scale, noble_metal_content, lifetime_months)
VALUES (?, 'CO', 'мг/м3', 200, 0.1626, 18)`, productTypeName)
	return merry.Wrap(err)
}

func GetLastParty(db *reform.DB, party *Party) error {
	err := db.SelectOneTo(party, `ORDER BY created_at DESC LIMIT 1;`)
	if err == reform.ErrNoRows {
		partyID, err := CreateNewParty(db)
		if err != nil {
			return merry.Wrap(err)
		}
		err = db.FindByPrimaryKeyTo(party, partyID)
	}
	if err != nil {
		return merry.Wrap(err)
	}
	party.Last = true
	return nil
}

func GetPartyProducts(db *reform.DB, party *Party) error {
	products, err := GetProductsInfoWithPartyID(db, party.PartyID)
	if err != nil {
		return merry.Wrap(err)
	}
	party.Products = products
	return nil
}

func GetPartyIsLast(db *reform.DB, party *Party) error {
	lastPartyID, err := GetLastPartyID(db)
	if err != nil {
		return merry.Wrap(err)
	}
	party.Last = party.PartyID == lastPartyID
	return nil
}

func CreateNewParty(db *reform.DB) (int64, error) {
	r, err := db.Exec(`INSERT INTO party DEFAULT VALUES`)
	if err != nil {
		return 0, merry.Wrap(err)
	}
	partyID, err := r.LastInsertId()
	if err != nil {
		return 0, merry.Wrap(err)
	}
	if r, err = db.Exec(`INSERT INTO product(party_id, serial, place) VALUES (?, 1, 0)`, partyID); err != nil {
		return 0, merry.Wrap(err)
	}
	logrus.Warnf("new party created: %d", partyID)
	return partyID, nil
}

func GetLastPartyID(db *reform.DB) (partyID int64, err error) {
	row := db.QueryRow(`SELECT party_id FROM party ORDER BY created_at DESC LIMIT 1`)
	err = row.Scan(&partyID)
	if err == sql.ErrNoRows {
		return CreateNewParty(db)
	}
	return partyID, merry.Wrap(err)
}

type ProductsFilter struct {
	WithSerials, WithProduction bool
}

func GetLastPartyProducts(db *reform.DB, f ProductsFilter) ([]Product, error) {
	tail := "WHERE party_id IN (SELECT party_id FROM last_party)"
	if f.WithSerials {
		tail += " AND (serial NOTNULL)"
	}
	if f.WithProduction {
		tail += " AND production"
	}
	xs, err := db.SelectAllFrom(ProductTable, tail)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	return structToProductSlice(xs), nil
}

func structToProductSlice(xs []reform.Struct) (products []Product) {
	for _, x := range xs {
		p := x.(*Product)
		products = append(products, *p)
	}
	return
}

func GetProductsInfoWithPartyID(db *reform.DB, partyID int64) ([]ProductInfo, error) {
	xs, err := db.SelectAllFrom(ProductInfoTable, "WHERE party_id = ? ORDER BY place", partyID)
	if err != nil {
		return nil, err
	}
	var productsInfo []ProductInfo
	for _, x := range xs {
		productsInfo = append(productsInfo, *x.(*ProductInfo))
	}
	return productsInfo, nil
}

func ListProductTypeNames(db *reform.DB) ([]string, error) {
	xs, err := db.SelectAllFrom(ProductTypeTable, "")
	if err != nil {
		return nil, err
	}
	var r []string
	for _, x := range xs {
		r = append(r, x.(*ProductType).ProductTypeName)
	}
	return r, nil
}

func ListUnits(db *reform.DB) ([]Units, error) {
	records, err := db.SelectAllFrom(UnitsTable, "")
	if err != nil {
		return nil, err
	}
	var units []Units
	for _, r := range records {
		x := r.(*Units)
		units = append(units, *x)
	}
	return units, nil
}

func ListGases(db *reform.DB) ([]Gas, error) {
	records, err := db.SelectAllFrom(GasTable, "")
	if err != nil {
		return nil, err
	}
	var gas []Gas
	for _, r := range records {
		x := r.(*Gas)
		gas = append(gas, *x)
	}
	return gas, nil
}

func GetLastPartyProductAtPlace(db *reform.DB, place int, product *Product) error {
	return db.SelectOneTo(product, "WHERE party_id = (SELECT party_id FROM last_party) AND place = ?", place)
}

func UpdateProductAtPlace(db *reform.DB, place int, f func(p *Product) error) (int64, error) {
	partyID, err := GetLastPartyID(db)
	if err != nil {
		return 0, nil
	}

	var p Product
	if err := db.SelectOneTo(&p, "WHERE party_id = ? AND place = ?", partyID, place); err != nil && err != reform.ErrNoRows {
		return 0, err
	}
	if err := f(&p); err != nil {
		return 0, err
	}
	p.PartyID = partyID
	p.Place = place
	if err := db.Save(&p); err != nil {
		return 0, err
	}
	return p.ProductID, nil
}
