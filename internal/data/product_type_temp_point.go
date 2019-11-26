package data

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/pkg/must"
	"strconv"
	"strings"
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

func GetProductTypeTempPoints(values []string, productTypeName string) ([]ProductTypeTempPoint, error) {
	if len(values)%3 != 0 {
		return nil, merry.New("sequence length is not a multiple of three")
	}

	var xs []ProductTypeTempPoint

	for n := 0; n < len(values); n += 3 {
		strT := strings.TrimSpace(values[n+0])
		if len(strT) == 0 {
			continue
		}

		r := ProductTypeTempPoint{ProductTypeName: productTypeName}
		var err error

		r.Temperature, err = parseFloat(values[n])
		if err != nil {
			return nil, merry.Appendf(err, "строка %d", n)
		}

		strI := strings.TrimSpace(values[n+1])
		r.Fon, err = parseFloatPtr(strI)
		if err != nil {
			return nil, merry.Appendf(err, "строка %d", n)
		}

		strS := strings.TrimSpace(values[n+2])
		r.Sens, err = parseFloatPtr(strS)
		if err != nil {
			return nil, merry.Appendf(err, "строка %d", n)
		}
		xs = append(xs, r)
	}
	return xs, nil
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.Replace(s, ",", ".", -1), 64)
}

func parseFloatPtr(s string) (*float64, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return nil, nil
	}
	s = strings.Replace(s, ",", ".", -1)
	v, err := strconv.ParseFloat(s, 64)
	return &v, err
}
