package data

import (
	"database/sql"
	"github.com/fpawel/elco/internal/elco"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/sqlite3"
)

//go:generate go run github.com/fpawel/goutils/dbutils/sqlstr/...

func Open(logger reform.Logger) (*reform.DB, error) {
	db, err := sql.Open("sqlite3", elco.DataFileName())
	if err != nil {
		return nil, err
	}
	dbr := reform.NewDB(db, sqlite3.Dialect, logger)
	_, err = dbr.Exec(SQLCreate)
	if err != nil {
		return nil, err
	}
	err = DeleteEmptyRecords(dbr)
	if err != nil {
		return nil, err
	}
	return dbr, nil
}

func DeleteEmptyRecords(dbr *reform.DB) error {
	_, err := dbr.Exec(`
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
	return err
}

func EnsureProductTypeName(db *reform.DB, productTypeName string) error {
	_, err := db.Exec(`
INSERT OR IGNORE INTO product_type 
  (product_type_name, gas_name, units_name, scale, noble_metal_content, lifetime_months)
VALUES (?, 'CO', 'мг/м3', 200, 0.1626, 18)`, productTypeName)
	return err
}

func GetLastParty(db *reform.DB) (Party, error) {
	var party Party
	err := db.SelectOneTo(&party, `ORDER BY created_at DESC LIMIT 1;`)
	if err == reform.ErrNoRows {
		partyID, err := CreateNewParty(db)
		if err == nil {
			err = db.FindByPrimaryKeyTo(&party, partyID)
		}
		return party, err
	}
	if err != nil {
		return party, err
	}
	party.Products, err = GetProductsInfoWithPartyID(db, party.PartyID)
	return party, err
}

func CreateNewParty(db *reform.DB) (int64, error) {
	r, err := db.Exec(`INSERT INTO party DEFAULT VALUES`)
	if err != nil {
		return 0, err
	}
	partyID, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}
	if r, err = db.Exec(`INSERT INTO product(party_id, serial, place) VALUES (?, 1, 0)`, partyID); err != nil {
		return 0, err
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
	return partyID, err
}

func DeletePartyID(db *reform.DB, partyID int64) error {
	_, err := db.Exec(`DELETE FROM party WHERE party_id = ?;`, &partyID)
	return err
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
		tail += " AND (production NOTNULL)"
	}
	rows, err := db.SelectRows(ProductTable, tail)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return fetchProductsFromRows(db, rows)
}

func GetProductsWithPartyID(db *reform.DB, partyID int64) ([]Product, error) {
	rows, err := db.SelectRows(ProductTable, "WHERE party_id = ?", partyID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return fetchProductsFromRows(db, rows)
}

func fetchProductsFromRows(db *reform.DB, rows *sql.Rows) ([]Product, error) {
	var products []Product
	for {
		var product Product
		if err := db.NextRow(&product, rows); err == nil {
			products = append(products, product)
		} else {
			if err == reform.ErrNoRows {
				return products, nil
			}
			return products, err
		}
	}
}

func GetProductsInfoWithPartyID(db *reform.DB, partyID int64) ([]ProductInfo, error) {
	var products []ProductInfo
	rows, err := db.SelectRows(ProductInfoTable, "WHERE party_id = ? ORDER BY place", partyID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	for {
		var product ProductInfo
		err = db.NextRow(&product, rows)
		if err == reform.ErrNoRows {
			return products, nil
		}
		if err != reform.ErrNoRows {
			return nil, err
		}
		products = append(products, product)
	}
}

func ListProductTypes(db *reform.DB) (prodTypes []ProductType) {

	rows, err := db.SelectRows(ProductTypeTable, "")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for {
		var pt ProductType
		if err = db.NextRow(&pt, rows); err != nil {
			break
		}
		prodTypes = append(prodTypes, pt)
	}
	if err != reform.ErrNoRows {
		panic(err)
	}
	return
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

func GetProductInfoAtPlace(db *reform.DB, place int) (ProductInfo, error) {
	party, err := GetLastParty(db)
	if err != nil {
		return ProductInfo{}, err
	}
	var p ProductInfo
	err = db.SelectOneTo(&p, "WHERE party_id = ? AND place = ?", party.PartyID, place)
	return p, err
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
