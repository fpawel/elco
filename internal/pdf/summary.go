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
	d.CellFormat(tableWidth, 6, tr(fmt.Sprintf("Итоговая таблица электрохимических ячеек №%d", party.PartyID)),
		"", 1, "C", false, 0, "")

	d.SetX(spaceX)
	y := d.GetY()
	d.SetFont("RobotoCondensed", "", 9)
	d.CellFormat(0, 6, tr(fmt.Sprintf("ИБЯЛ.418425.%s", party.ProductTypeName)),
		"", 1, "", false, 0, "")

	d.SetY(y)
	d.CellFormat(tableWidth, 6, tr(fmt.Sprintf("%s 0-%v %s",
		productType.GasName, productType.Scale, productType.UnitsName)),
		"", 1, "C", false, 0, "")

	d.SetY(y)
	d.CellFormat(tableWidth, 6, time.Now().Format("02.01.2006"),
		"", 1, "R", false, 0, "")

	d.SetFont("RobotoCondensed-Light", "", 8)

	colWidth1 := d.GetStringWidth("№ п/п")
	colWidth := (pageWidth - colWidth1 - 2.*spaceX) / 9.

	for i, str := range []string{
		"№ п/п",
		"Зав.№",
		"Iфон, мкА",
		"Dфон, мкА",
		"Dt, мкА",
		fmt.Sprintf("Кч20, мкА/%s", productType.UnitsName),
		"Кч50, %",
		fmt.Sprintf("Dn, мкА/%s", productType.UnitsName),
		"Uсоу, мВ",
		"Uстг, мВ",
	} {
		w := colWidth
		if i == 0 {
			w = colWidth1
		}
		d.CellFormat(w, 6, tr(str), "1", 0, "C", true, 0, "")
	}
	d.Ln(-1)

	for row, p := range party.Products {
		d.CellFormat(colWidth1, 3.5, strconv.Itoa(row+1), "1", 0, "C", true, 0, "")

		d.CellFormat(colWidth, 3.5, fmt.Sprintf("%d-%d", p.Serial.Int64, p.ProductID),
			"1", 0, "C", false, 0, "")

		d.CellFormat(colWidth, 3.5, formatNullFloat64(p.IFPlus20), "1", 0, "R", false, 0, "")
		d.CellFormat(colWidth, 3.5, formatNullFloat64(p.DFon20), "1", 0, "R", false, 0, "")
		d.CellFormat(colWidth, 3.5, formatNullFloat64(p.DFon50), "1", 0, "R", false, 0, "")
		d.CellFormat(colWidth, 3.5, formatNullFloat64(p.KSens20), "1", 0, "R", false, 0, "")
		d.CellFormat(colWidth, 3.5, formatNullFloat64(p.KSens50), "1", 0, "R", false, 0, "")
		d.CellFormat(colWidth, 3.5, formatNullFloat64(p.DNotMeasured), "1", 0, "R", false, 0, "")

		var u1, u2 string
		if p.DFon50.Valid {
			u1 = strconv.FormatFloat(p.DFon50.Float64*77, 'f', 1, 64)
			u2 = strconv.FormatFloat(p.DFon50.Float64*52, 'f', 1, 64)
		}
		d.CellFormat(colWidth, 3.5, u1, "1", 0, "R", false, 0, "")
		d.CellFormat(colWidth, 3.5, u2, "1", 0, "R", false, 0, "")
		d.Ln(-1)
	}
}
