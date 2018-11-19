package data

import (
	"github.com/jmoiron/sqlx"
	"gopkg.in/reform.v1"
)

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

func GetProductsByPartyID(db *reform.DB, partyID int64) (products []ProductInfo) {
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
