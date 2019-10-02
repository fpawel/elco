package api

import (
	"database/sql"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/pkg"
	"strconv"
)

type ProductsCatalogueSvc struct{}

type Cell struct {
	Str           string
	Res           ValueResult
	TextAlignment TextAlignment
}

type ValueResult int

const (
	_ ValueResult = iota
	vrOk
	vrErr
)

type TextAlignment int

const (
	taLeftJustify TextAlignment = iota
	taRightJustify
	taCenter
)

func cell1(s string, ta TextAlignment) Cell {
	return Cell{Str: s, TextAlignment: ta}
}

func cell2(v sql.NullFloat64) Cell {
	var c Cell
	if !v.Valid {
		return c
	}
	c.Str = pkg.FormatFloat(v.Float64, 3)
	c.TextAlignment = taRightJustify
	return c
}

func cellR2(v sql.NullFloat64, f1, f2 bool) Cell {
	var c Cell
	if !v.Valid {
		return c
	}
	c.Str = pkg.FormatFloat(v.Float64, 3)
	c.TextAlignment = taRightJustify
	if f1 && f2 {
		c.Res = vrOk
	} else {
		c.Res = vrErr
	}
	return c
}

func cellS(s sql.NullString) Cell {
	var c Cell
	if !s.Valid || len(s.String) == 0 {
		return c
	}
	c.Str = s.String
	c.TextAlignment = taLeftJustify
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
		"Примечание",
		"Примечание загрузки",
	}
	r = make([]Cell, len(xs))
	for i := range r {
		r[i] = cell1(xs[i], taCenter)
	}
	r[len(r)-2].TextAlignment = taLeftJustify
	r[len(r)-1].TextAlignment = taLeftJustify

	return r
}

func productsTable(products []data.ProductInfo) [][]Cell {
	r1 := [][]Cell{
		products3HeaderRow(),
	}
	cols := map[int]struct{}{}
	for _, p := range products {
		row := []Cell{
			cell1(strconv.Itoa(int(p.ProductID)), taCenter),
			cell1(p.CreatedAt.Format("02.01.2006"), taCenter),
			cell1(strconv.FormatBool(p.HasFirmware), taCenter),
			cell1(strconv.Itoa(int(p.PartyID)), taCenter),
			cell1(strconv.Itoa(int(p.Serial.Int64)), taCenter),
			cell1(data.FormatPlace(p.Place), taCenter),
			cell1(p.AppliedProductTypeName, taCenter),

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
			cellS(p.NoteProduct),
			cellS(p.NoteParty),
		}
		r1 = append(r1, row)
		for i, c := range row {
			if len(c.Str) > 0 {
				cols[i] = struct{}{}
			}
		}
	}
	r2 := make([][]Cell, 0)
	for _, row1 := range r1 {
		var row2 []Cell
		for col, cell := range row1 {
			if _, f := cols[col]; f {
				row2 = append(row2, cell)
			}
		}
		r2 = append(r2, row2)
	}
	return r2
}

func (_ *ProductsCatalogueSvc) ProductByID(productID [1]int64, r *[][]Cell) error {
	return fetchProducts(r, "WHERE product_id = ?", productID[0])
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
