package pdf

import (
	"fmt"
	"github.com/fpawel/elco/internal/data"
	"github.com/jung-kurt/gofpdf"
)

func passportSou(dir string, party data.Party) error {

	d, err := newDoc()
	if err != nil {
		return err
	}

	pageWidth, _ := d.GetPageSize()
	const spaceX = 10.
	width := pageWidth/2. - spaceX*2
	for i := range party.Products {
		if i%2 > 0 {
			continue
		}
		if i%8 == 0 {
			d.AddPage()
		} else {

			d.SetLineWidth(0.2)
			d.SetDrawColor(0, 0, 0)

			d.MoveTo(d.GetX(), d.GetY()+5)
			d.LineTo(pageWidth-spaceX, d.GetY())
			d.DrawPath("D")

			d.SetLineWidth(.1)
			d.SetDrawColor(169, 169, 169)

			d.Ln(5)
		}
		y := d.GetY()
		doPassportSou(d, spaceX, width, party.Products[i])
		d.SetY(y)
		if i == len(party.Products)-1 {
			break
		}
		doPassportSou(d, pageWidth/2., width, party.Products[i+1])
	}

	if err := saveAndShowDoc(d, dir, "sou"); err != nil {
		return err
	}

	return nil
}

func doPassportSou(d *gofpdf.Fpdf, left, width float64, p data.ProductInfo) {

	d.SetX(left)

	tr := d.UnicodeTranslatorFromDescriptor("cp1251")
	sentence := func(familyStr, fontStyleStr string, fontSize float64, h float64, s string) {
		d.SetFont(familyStr, fontStyleStr, fontSize)
		d.CellFormat(d.GetStringWidth(tr(s)), h, tr(s), "", 0, "", false, 0, "")
	}

	sentence1 := func(s string) {
		sentence("RobotoCondensed-Light", "", fontSize1, lineSpace1, s)
	}
	sentenceB := func(s string) {
		sentence("RobotoCondensed", "B", fontSize1, lineSpace1, s)
	}

	passportPart1(d, left, width, p)

	y := d.GetY()
	x := 13 + 8 + 8 + 0.3

	d.SetY(d.GetY() + 1)
	colWidths := []float64{8, 13, 8}

	for row, c := range [][]string{
		{"T, \"C", "Iфон, мкА", "Кч, %"},
		{"20", formatNullFloat64(p.IFPlus20), "100"},
		{"50", formatNullFloat64(p.IFPlus50), formatNullFloat64(p.KSens50)},
	} {
		d.SetX(left)
		for col, str := range c {
			f := col == 0 || row == 0
			align := "R"
			if f {
				align = "C"
				d.SetFont("RobotoCondensed-Light", "", 8)
			} else {
				d.SetFont("RobotoCondensed-Light", "", 7)
			}
			d.CellFormat(colWidths[col], 3.5, tr(str), "1", 0, align, f, 0, "")
		}
		d.Ln(-1)
	}

	d.SetY(y)
	d.SetX(left + x)

	d.SetFont("RobotoCondensed-Light", "", fontSize1)
	d.CellFormat(0, 5, tr("Содержание драгоценных металлов:"),
		"", 1, "", false, 0, "")

	d.SetX(left + x)
	sentence1("платина ")
	sentenceB(fmt.Sprintf("%v г.", p.NobleMetalContent))
	sentence1(" Ячейка соответст-")
	d.Ln(lineSpace1)

	d.SetFont("RobotoCondensed-Light", "", fontSize1)
	d.SetX(left + x)
	d.CellFormat(0, 4, tr("вует комплекту документации"),
		"", 1, "", false, 0, "")

	d.SetY(d.GetY() + 0.3)
	d.SetX(left)

	d.MultiCell(width, 4, tr(fmt.Sprintf(`ИБЯЛ.418425.%s и признана годной к эксплуатации. Гарантийный срок эксплуатации со дня отгрузки %d месяцев, но не более 18 месяцев со дня изготовления.`,
		p.AppliedProductTypeName, p.LifetimeMonths)),
		"", "", false)
	d.Ln(2)
	d.SetX(left + 50)
	d.SetFont("RobotoCondensed", "B", 9)
	d.CellFormat(0, 4, tr("ОТК:"),
		"", 1, "", false, 0, "")
}

func passportPart1(d *gofpdf.Fpdf, left, width float64, p data.ProductInfo) {

	tr := d.UnicodeTranslatorFromDescriptor("cp1251")
	sentence := func(familyStr, fontStyleStr string, fontSize float64, h float64, s string) {
		d.SetFont(familyStr, fontStyleStr, fontSize)
		d.CellFormat(d.GetStringWidth(tr(s)), h, tr(s), "", 0, "", false, 0, "")
	}

	sentence1 := func(s string) {
		sentence("RobotoCondensed-Light", "", fontSize1, lineSpace1, s)
	}
	sentenceB := func(s string) {
		sentence("RobotoCondensed", "B", fontSize1, lineSpace1, s)
	}

	d.SetX(left)
	d.SetFont("RobotoCondensed", "B", 11)
	d.CellFormat(width, 9, tr("Электрохимическая ячейка ИБЯЛ.418425."+p.AppliedProductTypeName),
		"", 1, "", false, 0, "")

	d.SetX(left)
	sentence1("Дата: ")
	sentenceB(p.CreatedAt.Format("02.01.2006"))
	sentence1(" Зав.номер: ")
	sentenceB(fmt.Sprintf("%s-%d", formatNullInt64(p.Serial), p.ProductID))
	d.Ln(lineSpace1)

	d.SetX(left)
	sentence1(fmt.Sprintf("Коэффициент чувствительности, мкА/%s: ", p.UnitsName))
	sentenceB(formatNullFloat64(p.KSens20))
	d.Ln(lineSpace1)

	d.SetX(left)
	sentence1("Измеряемый компонент: ")
	sentenceB(p.GasName + " ")
	sentence1(fmt.Sprintf("Диапазон, мкА/%s: ", p.UnitsName))
	sentenceB(fmt.Sprintf("0-%v", p.Scale))
	d.Ln(lineSpace1)
}
