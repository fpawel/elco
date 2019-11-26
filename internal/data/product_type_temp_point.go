package data

import (
	"github.com/fpawel/elco/internal/pkg/must"
)

type ProductTypeTempPoint struct {
	Temperature     float64  `db:"temperature"`
	Fon             *float64 `db:"fon"`
	Sens            *float64 `db:"sens"`
	ProductTypeName string   `db:"product_type_name"`
}

func FetchProductTypeTempPoints(productTypeName string) (productTypeTempPoints []ProductTypeTempPoint) {
	err := DBx.Select(&productTypeTempPoints,
		`SELECT temperature, fon, sens, product_type_name FROM product_type_current WHERE product_type_name=? ORDER BY temperature`,
		productTypeName)
	must.PanicIf(err)
	return
}
