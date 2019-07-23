package pdf

import (
	"fmt"
	"github.com/fpawel/elco/internal/data"
	"github.com/jung-kurt/gofpdf"
	"strconv"
	"time"
)

func summary(dir string, party data.Party) error {

	var productType data.ProductType
	if err := data.DB.FindByPrimaryKeyTo(&productType, party.ProductTypeName); err != nil {
		return err
	}

	d, err := newDoc()
	if err != nil {
		return err
	}

	doSummary(d, party, productType)

	if err := saveAndShowDoc(d, dir, "summary"); err != nil {
		return err
	}

	return nil
}

func doSummary(d *gofpdf.Fpdf, party data.Party, productType data.ProductType) {
	tr := d.UnicodeTranslatorFromDescriptor("cp1251")

	d.AddPage()

	const spaceX = 10.
	pageWidth, _ := d.GetPageSize()
	tableWidth := pageWidth - 2.*spaceX

	d.SetX(spaceX)
	d.SetFont("RobotoCondensed", "B", 13)
	d.CellFormat(tableWidth, 6, tr(fmt.Sprintf(
		"Итоговая таблица электрохимических ячеек №%d", party.PartyID)),
		"", 1, "C", false, 0, "")

	d.SetX(spaceX)
	d.SetFont("RobotoCondensed", "", 9)
	d.CellFormat(0, 6,
		tr(fmt.Sprintf("ИБЯЛ.418425.%s, %s 0-%v %s, %d штук",
			party.ProductTypeName,
			productType.GasName, productType.Scale, productType.UnitsName,
			len(party.Products)+1)),
		"", 0, "", false, 0, "")

	d.SetX(spaceX)
	d.CellFormat(tableWidth, 6, time.Now().Format("02.01.2006"),
		"", 1, "R", false, 0, "")

	d.SetFont("RobotoCondensed-Light", "", 9.5)

	colWidth1 := d.GetStringWidth("№ п/п")
	colWidth := (pageWidth - colWidth1 - 2.*spaceX) / 9.

	for i, str := range []string{
		"№ п/п",
		"Зав.№",
		"Iфон, мкА",
		"Dф50, мкА",
		"Кч-20, %",
		"Kч50, %",
		"Uсоу, мВ",
		"Uстг, мВ",
		"Код 1",
		"Код 2",
	} {
		w := colWidth
		if i == 0 {
			w = colWidth1
		}
		f := i == 5 || i == 7
		if f {
			d.SetFont("RobotoCondensed-Light", "", 8)
		}

		borderStr := "LT"
		if i > 0 {
			borderStr += "R"
		}

		d.CellFormat(w, 6, tr(str), "LRT", 0, "C", false, 0, "")
		if f {
			d.SetFont("RobotoCondensed-Light", "", 9.5)
		}
	}
	d.Ln(-1)

	const collHeight = 4
	cf := func(txtStr, borderStr string, ln int, alignStr string) {
		d.CellFormat(colWidth, collHeight, txtStr, borderStr, ln, alignStr, false, 0, "")
	}
	for row, p := range party.Products {
		borderStr := "TL"
		if row == len(party.Products)-1 {
			borderStr += "B"
		}
		d.CellFormat(colWidth1, collHeight, strconv.Itoa(row+1), borderStr,
			0, "C", false, 0, "")
		cf(fmt.Sprintf("%d-%d", p.Serial.Int64, p.ProductID), borderStr, 0, "C")

		cf(formatNullFloat64(p.IFPlus20), borderStr, 0, "R")
		cf(formatNullFloat64(p.DFon50), borderStr, 0, "R")
		cf(formatNullFloat64(p.KSensMinus20), borderStr, 0, "R")
		cf(formatNullFloat64(p.KSens50), borderStr, 0, "R")

		var u1, u2 string
		if p.DFon50.Valid {
			u1 = strconv.FormatFloat(p.DFon50.Float64*77, 'f', 1, 64)
			u2 = strconv.FormatFloat(p.DFon50.Float64*52, 'f', 1, 64)
		}
		cf(u1, borderStr, 0, "R")
		cf(u2, borderStr, 0, "R")

		code1, code2 := tempCodes(p)
		cf(code1, borderStr, 0, "C")
		cf(code2, borderStr+"R", 0, "C")
		d.Ln(-1)
	}
}
