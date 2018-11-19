package crud

import (
	"github.com/fpawel/elco/internal/data"
	"gopkg.in/reform.v1"
)

type ProductTypes struct {
	dbContext
}

func (x ProductTypes) Names() (names []string) {
	x.mu.Lock()
	defer x.mu.Unlock()

	rows, err := x.dbr.SelectRows(data.ProductTypeTable, "")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for {
		var pt data.ProductType
		if err = x.dbr.NextRow(&pt, rows); err != nil {
			break
		}
		names = append(names, pt.ProductTypeName)
	}
	if err != reform.ErrNoRows {
		panic(err)
	}
	return
}
