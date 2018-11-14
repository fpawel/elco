package data

import "gopkg.in/reform.v1"

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
