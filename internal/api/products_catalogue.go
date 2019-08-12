package api

import (
	"database/sql"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/gohelp/myfmt"
	"strconv"
)

type ProductsCatalogueSvc struct{}

type Cell struct {
	Str string
	Res ValueResult
}

type ValueResult int

const (
	vrNone ValueResult = iota
	vrOk
	vrErr
)

func cell1(s string) Cell {
	return Cell{Str: s}
}

func cell2(v sql.NullFloat64) Cell {
	var c Cell
	if !v.Valid {
		return c
	}
	c.Str = myfmt.FormatFloat(v.Float64, 3)
	return c
}

func cellR2(v sql.NullFloat64, f1, f2 bool) Cell {
	var c Cell
	if !v.Valid {
		return c
	}
	c.Str = myfmt.FormatFloat(v.Float64, 3)
	if f1 && f2 {
		c.Res = vrOk
	} else {
		c.Res = vrErr
	}
	return c
}

func products3HeaderRow() (r []Cell) {
	xs := []string{
		"ID",
		"Дата",
		"Прошивка",
		"Загрузка",
		"Зав.№",
		"Место",
		"Исполнение",
		"ФОН20",
		"Ч20",
		"Kч20",
		"ФОН20.2",
		"ПГС2", "ПГС3", "ПГС2.2", "ПГС1", "неизм.",
		"ФОН-20", "Ч-20", "Kч-20",
		"ФОН50", "Ч50", "Kч50",
	}
	r = make([]Cell, len(xs))
	for i := range r {
		r[i] = cell1(xs[i])
	}

	return nil
}

func productsTable(products []data.ProductInfo) [][]Cell {
	r1 := [][]Cell{
		products3HeaderRow(),
	}
	cols := map[int]struct{}{}
	for _, p := range products {
		row := []Cell{
			cell1(strconv.Itoa(int(p.ProductID))),
			cell1(p.CreatedAt.Format("02.01.2006")),
			cell1(strconv.FormatBool(p.HasFirmware)),
			cell1(strconv.Itoa(int(p.PartyID))),
			cell1(strconv.Itoa(int(p.Serial.Int64))),
			cell1(data.FormatPlace(p.Place)),
			cell1(p.AppliedProductTypeName),

			cellR2(p.IFPlus20, p.OKMinFon20, p.OKMaxFon20),
			cellR2(p.ISPlus20, p.OKMinKSens20, p.OKMaxKSens20),
			cellR2(p.KSens20, p.OKMinKSens20, p.OKMaxKSens20),

			cellR2(p.I13, p.OKMinFon20r, p.OKMaxFon20r),
			cell2(p.I24),
			cell2(p.I35),
			cell2(p.I26),
			cell2(p.I17),
			cell2(p.NotMeasured),

			cell2(p.IFMinus20),
			cell2(p.ISMinus20),
			cell2(p.KSensMinus20),

			cellR2(p.IFPlus50, p.OKDFon50, p.OKDFon50),
			cellR2(p.ISPlus50, p.OKMinKSens50, p.OKMaxKSens50),
			cellR2(p.KSens50, p.OKMinKSens50, p.OKMaxKSens50),
		}
		r1 = append(r1, row)
		for i, c := range row {
			if len(c.Str) > 0 {
				cols[i] = struct{}{}
			}
		}
	}
	var r2 [][]Cell
	for _, row1 := range r1 {
		var row2 []Cell
		for col, cell := range row1 {
			if _, f := cols[col]; !f {
				row2 = append(row2, cell)
			}
		}
		r2 = append(r2, row2)
	}
	return r2
}

func (_ *ProductsCatalogueSvc) ListProductsBySerial(serial [1]int, r *[][]Cell) error {
	var products []data.ProductInfo
	xs, err := data.DB.SelectAllFrom(data.ProductInfoTable, "WHERE serial = ?", serial[0])
	if err != nil {
		return err
	}
	for _, p := range xs {
		products = append(products, *p.(*data.ProductInfo))
	}
	*r = productsTable(products)

	return nil
}
