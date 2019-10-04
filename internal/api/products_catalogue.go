package api

import (
	"github.com/fpawel/elco/internal/data"
)

type ProductsCatalogueSvc struct{}

func (_ *ProductsCatalogueSvc) ProductByIDHasFirmware(productID [1]int64, r *bool) error {
	return data.DBx.Get(nil, `SELECT has_firmware FROM product_info WHERE product_id = ?`, productID[0])
}

func (_ *ProductsCatalogueSvc) ProductByID(productID [1]int64, r *[][]Cell) error {
	return fetchProducts(r, "WHERE product_id = ?", productID[0])
}

func (_ *ProductsCatalogueSvc) ProductInfoByID(productID [1]int64, p *data.ProductInfo) error {
	return data.DB.FindByPrimaryKeyTo(p, productID[0])
}

func (_ *ProductsCatalogueSvc) ListProductsByPartyID(partyID [1]int, r *[][]Cell) error {
	return fetchProducts(r, "WHERE party_id = ? ORDER BY place DESC", partyID[0])
}

func (_ *ProductsCatalogueSvc) ListProductsBySerial(serial [1]int, r *[][]Cell) error {
	return fetchProducts(r, "WHERE serial = ? ORDER BY created_at DESC", serial[0])
}

func (_ *ProductsCatalogueSvc) ListProductsByNote(note [1]string, r *[][]Cell) error {
	return fetchProducts(r,
		"WHERE note_product LIKE $1 OR note_party LIKE $1 ORDER BY created_at DESC LIMIT 1000",
		"%"+note[0]+"%")
}

func fetchProducts(r *[][]Cell, tail string, args ...interface{}) error {
	xs, err := data.DB.SelectAllFrom(data.ProductInfoTable, tail, args...)
	if err != nil {
		return err
	}
	products := make([]data.ProductInfo, 0)
	for _, p := range xs {
		products = append(products, *p.(*data.ProductInfo))
	}
	*r = productsTable(products)
	return nil
}
