package api

import (
	"database/sql"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/pkg"
	"strconv"
	"strings"
)

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

func productsTitles(xs1, xs2 []string) []string {
	return append(xs1, append([]string{
		"ФОН20",
		"Ч20",
		"Kч20",
		"ФОН20.2",
		"ПГС2", "ПГС3", "ПГС2.2", "ПГС1", "неизм.",
		"ФОН-20", "Ч-20", "Kч-20",
		"ФОН50", "Ч50", "Kч50",
	}, xs2...)...)
}

func productsTable2(products []data.ProductInfo) [][]Cell {

	var row0 []Cell

	titles := productsTitles([]string{
		"№",
		"Зав.№",
		"ID",
		"Прошивка",
	}, []string{
		"Иполнение",
		"Т.расчёт",
		"Примечание",
	})

	for _, s := range titles {
		ta := taCenter
		if strings.HasPrefix(s, "Примечание") {
			ta = taLeftJustify
		}
		row0 = append(row0, cellStr(s, ta))
	}
	rows := [][]Cell{row0}
	cols := map[int]struct{}{}
loop96:
	for i := 0; i < 96; i++ {
		for _, p := range products {
			if p.Place != i {
				continue
			}
			rows = appendProductValuesRow(rows, cols, productValuesCells([]Cell{
				cellStr(data.FormatPlace(p.Place), taCenter),
				cellStr(strconv.Itoa(int(p.Serial.Int64)), taCenter),
				cellStr(strconv.Itoa(int(p.ProductID)), taCenter),
				cellStr(strconv.FormatBool(p.HasFirmware), taCenter),
			}, []Cell{
				cellNullStr(p.ProductTypeName),
				cellNullInt(p.PointsMethod),
				cellNullStr(p.NoteProduct),
			}, p))
			continue loop96
		}
		rows = append(rows, []Cell{cellStr(data.FormatPlace(i), taCenter)})

	}
	return removeEmptyCols(rows, cols)
}

func productsTable(products []data.ProductInfo) [][]Cell {

	var row0 []Cell

	titles := productsTitles([]string{
		"ID",
		"Дата",
		"Прошивка",
		"Загрузка",
		"Зав.№",
		"Место",
		"Исполнение",
	}, []string{
		"Примечание",
		"Примечание загрузки",
	})

	for _, s := range titles {
		ta := taCenter
		if strings.HasPrefix(s, "Примечание") {
			ta = taLeftJustify
		}
		row0 = append(row0, cellStr(s, ta))
	}
	rows := [][]Cell{row0}

	cols := map[int]struct{}{}
	for _, p := range products {
		rows = appendProductValuesRow(rows, cols, productValuesCells([]Cell{
			cellStr(strconv.Itoa(int(p.ProductID)), taCenter),
			cellStr(p.CreatedAt.Format("02.01.2006"), taCenter),
			cellStr(strconv.FormatBool(p.HasFirmware), taCenter),
			cellStr(strconv.Itoa(int(p.PartyID)), taCenter),
			cellStr(strconv.Itoa(int(p.Serial.Int64)), taCenter),
			cellStr(data.FormatPlace(p.Place), taCenter),
			cellStr(p.AppliedProductTypeName, taCenter),
		}, []Cell{
			cellNullStr(p.NoteProduct),
			cellNullStr(p.NoteParty),
		}, p))
	}
	return removeEmptyCols(rows, cols)
}

func appendProductValuesRow(rows [][]Cell, cols map[int]struct{}, row []Cell) [][]Cell {
	for i, c := range row {
		if len(c.Str) > 0 {
			cols[i] = struct{}{}
		}
	}
	return append(rows, row)
}

func productValuesCells(xs1, xs2 []Cell, p data.ProductInfo) []Cell {
	xs1 = append(xs1, cellNullFloatCheck2(p.IFPlus20, p.OKMinFon20, p.OKMaxFon20),
		cellNullFloatCheck2(p.ISPlus20, p.OKMinKSens20, p.OKMaxKSens20),
		cellNullFloatCheck2(p.KSens20, p.OKMinKSens20, p.OKMaxKSens20),

		cellNullFloatCheck2(p.I13, p.OKMinFon20r, p.OKMaxFon20r),
		cellNullFloat(p.I24),
		cellNullFloat(p.I35),
		cellNullFloat(p.I26),
		cellNullFloat(p.I17),
		cellNullFloat(p.NotMeasured),

		cellNullFloat(p.IFMinus20),
		cellNullFloat(p.ISMinus20),
		cellNullFloat(p.KSensMinus20),

		cellNullFloatCheck2(p.IFPlus50, p.OKDFon50, p.OKDFon50),
		cellNullFloatCheck2(p.ISPlus50, p.OKMinKSens50, p.OKMaxKSens50),
		cellNullFloatCheck2(p.KSens50, p.OKMinKSens50, p.OKMaxKSens50))
	return append(xs1, xs2...)
}

func cellStr(s string, ta TextAlignment) Cell {
	return Cell{Str: s, TextAlignment: ta}
}

func cellNullFloat(v sql.NullFloat64) Cell {
	var c Cell
	if !v.Valid {
		return c
	}
	c.Str = pkg.FormatFloat(v.Float64, 3)
	c.TextAlignment = taRightJustify
	return c
}

func cellNullInt(v sql.NullInt64) Cell {
	var c Cell
	if !v.Valid {
		return c
	}
	c.Str = strconv.Itoa(int(v.Int64))
	c.TextAlignment = taRightJustify
	return c
}

func cellNullFloatCheck2(v sql.NullFloat64, f1, f2 bool) Cell {
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

func cellNullStr(s sql.NullString) Cell {
	var c Cell
	if !s.Valid || len(s.String) == 0 {
		return c
	}
	c.Str = s.String
	c.TextAlignment = taLeftJustify
	return c
}

func removeEmptyCols(rows [][]Cell, cols map[int]struct{}) [][]Cell {
	r := make([][]Cell, 0)
	for _, row1 := range rows {
		var row2 []Cell
		for col, cell := range row1 {
			if _, f := cols[col]; f {
				row2 = append(row2, cell)
			}
		}
		r = append(r, row2)
	}
	return r
}
