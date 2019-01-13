package data

import (
	"database/sql"
	"github.com/fpawel/goutils/dbutils"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"gopkg.in/reform.v1"
)

//go:generate go run github.com/fpawel/goutils/dbutils/sqlstr/...

func DeleteEmptyRecords(db *sqlx.DB) {
	db.MustExec(`
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

}

func EnsureParty(db *sqlx.DB) {
	var v bool
	dbutils.MustGet(db, &v, `SELECT exists(SELECT * FROM party LIMIT 1);`)
	if !v {
		CreateNewParty(db)
	}
}

func EnsureProductTypeName(db *sqlx.DB, productTypeName string) {
	db.MustExec(`
INSERT OR IGNORE INTO product_type 
  (product_type_name, gas_name, units_name, scale, noble_metal_content, lifetime_months)
VALUES (?, 'CO', 'мг/м3', 200, 0.1626, 18)`, productTypeName)
}

func GetLastParty(db *reform.DB) (party Party, err error) {
	if err = db.SelectOneTo(&party, `ORDER BY created_at DESC LIMIT 1;`); err != nil {
		return
	}
	party.Products = GetProductsInfoByPartyID(db, party.PartyID)
	return
}

func MustLastParty(db *reform.DB) Party {
	if party, err := GetLastParty(db); err != nil {
		panic(err)
	} else {
		return party
	}
}

func CreateNewParty(db *sqlx.DB) {
	r := db.MustExec(`INSERT INTO party DEFAULT VALUES`)
	partyID, err := r.LastInsertId()
	if err != nil {
		panic(err)
	}

	r = db.MustExec(`INSERT INTO product(party_id, serial, place) VALUES (?, 1, 0)`, partyID)
	if _, err = r.LastInsertId(); err != nil {
		panic(err)
	}
	logrus.WithField("party_id", partyID).Warn("new party created")
}

func LastPartyID(db *sqlx.DB) (partyID int64) {
	dbutils.MustGet(db, &partyID, `SELECT party_id FROM last_party;`)
	return
}

func DeletePartyID(db *sqlx.DB, partyID int64) error {
	_, err := db.Exec(`DELETE FROM party WHERE party_id = ?;`, &partyID)
	return err
}

func GetLastPartyProductsWithSerials(db *reform.DB) []Product {
	rows, err := db.SelectRows(ProductTable,
		"WHERE party_id IN (SELECT party_id FROM last_party) AND (serial NOTNULL)")
	if err != nil {
		panic(err)
	}
	defer func() { _ = rows.Close() }()
	return fetchProductsFromRows(db, rows)
}

func GetLastPartyProductsWithProduction(db *reform.DB) []Product {
	rows, err := db.SelectRows(ProductTable,
		"WHERE party_id IN (SELECT party_id FROM last_party) AND (serial NOTNULL)")
	if err != nil {
		panic(err)
	}
	defer func() { _ = rows.Close() }()
	return fetchProductsFromRows(db, rows)
}

func GetLastPartyProducts(db *reform.DB) []Product {
	rows, err := db.SelectRows(ProductTable,
		"WHERE party_id IN (SELECT party_id FROM last_party)")
	if err != nil {
		panic(err)
	}
	defer func() { _ = rows.Close() }()
	return fetchProductsFromRows(db, rows)
}

func fetchProductsFromRows(db *reform.DB, rows *sql.Rows) (products []Product) {
	for {
		var product Product
		if err := db.NextRow(&product, rows); err == nil {
			products = append(products, product)
		} else {
			if err == reform.ErrNoRows {
				return products
			}
			panic(err)
		}
	}
}

func GetProductsInfoByPartyID(db *reform.DB, partyID int64) (products []ProductInfo) {
	rows, err := db.SelectRows(ProductInfoTable, "WHERE party_id = ? ORDER BY place",
		partyID)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for {
		var product ProductInfo
		if err = db.NextRow(&product, rows); err != nil {
			break
		}
		products = append(products, product)
	}
	if err != reform.ErrNoRows {
		panic(err)
	}
	return
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

func ListUnits(db *reform.DB) (units []Units) {
	rows, err := db.SelectRows(UnitsTable, "")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for {
		var unit Units
		if err = db.NextRow(&unit, rows); err != nil {
			break
		}
		units = append(units, unit)
	}
	if err != reform.ErrNoRows {
		panic(err)
	}
	return
}

func ListGases(db *reform.DB) (gases []Gas) {
	rows, err := db.SelectRows(GasTable, "")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for {
		var gas Gas
		if err = db.NextRow(&gas, rows); err != nil {
			break
		}
		gases = append(gases, gas)
	}
	if err != reform.ErrNoRows {
		panic(err)
	}
	return
}
