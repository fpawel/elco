package pdf

import (
	"database/sql"
	"fmt"
	"github.com/fpawel/elco/internal/data"
	"github.com/jung-kurt/gofpdf"
	"strconv"
)

func passportDax(dir string, party data.Party) error {

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
		if i%6 == 0 {
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
		doPassportDax(d, spaceX, width, party.Products[i], party)
		d.SetY(y)
		if i == len(party.Products)-1 {
			break
		}
		doPassportDax(d, pageWidth/2., width, party.Products[i+1], party)

	}

	if err := saveAndShowDoc(d, dir, "dax"); err != nil {
		return err
	}

	return nil
}

func doPassportDax(d *gofpdf.Fpdf, left, width float64, p data.ProductInfo, party data.Party) {

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

	d.SetX(left)
	d.SetFont("RobotoCondensed", "BI", 7)
	d.CellFormat(0, lineSpace1, tr("Температурная зависимость фонового тока и чувствительности:"),
		"", 1, "", false, 0, "")

	f := func(float64) string {
		return ""
	}
	fFon, fSens := f, f
	if t, err := p.TableFon(); err == nil {
		a := data.NewApproximationTable(t)
		fFon = func(x float64) string {
			return strconv.FormatFloat(a.F(x), 'f', 3, 64)
		}
	}
	if t, err := p.TableSens(); err == nil {
		a := data.NewApproximationTable(t)
		fSens = func(x float64) string {
			return strconv.FormatFloat(a.F(x), 'f', 0, 64)
		}
	}

	dax := [][]string{
		{"T,\"C", "-20", "0", "20", "30", "50"},
		{"Iфон, мкА",
			formatNullFloat64Prec(p.IFMinus20, 3),
			fFon(0),
			formatNullFloat64Prec(p.IFPlus20, 3),
			fFon(30),
			formatNullFloat64Prec(p.IFPlus50, 3),
		},
		{"Кч, %",
			fSens(-20),
			fSens(0),
			fSens(20),
			fSens(30),
			fSens(50),
		},
	}

	colWidths := []float64{13, 13, 13, 13, 13, 13}

	renderTable := func() {
		for row, c := range dax {
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
				d.CellFormat(colWidths[col], 3.5, tr(str), "1", 0, align, false, 0, "")
			}
			d.Ln(-1)
		}
	}

	renderTable()

	dax = [][]string{
		{"", "ПГС1", "ПГС3", "ПГС1", "ПГС2", "ПГС3", "ПГС2", "ПГС1", "B"},
		{"мкА",
			formatNullFloat64Prec(p.IFPlus20, 3),
			formatNullFloat64Prec(p.ISPlus20, 3),
			formatNullFloat64Prec(p.I13, 3),
			formatNullFloat64Prec(p.I24, 3),
			formatNullFloat64Prec(p.I35, 3),
			formatNullFloat64Prec(p.I26, 3),
			formatNullFloat64Prec(p.I17, 3),
			formatNullFloat64Prec(p.Variation, 2),
		},
		{"мг/м3", "", "", "", "", "", "", "", ""},
	}
	colWidths = []float64{9, 9.5, 9.5, 9.5, 9.5, 9.5, 9.5, 9.5, 9.5}

	fon20, sens20 := p.IFPlus20.Float64, p.ISPlus20.Float64

	if fon20 != sens20 {

		dK := (party.Concentration3 - party.Concentration1) / (sens20 - fon20)

		for i, x := range []struct {
			sql.NullFloat64
			PGS float64
		}{
			{p.I13, party.Concentration1},
			{p.I24, party.Concentration2},
			{p.I35, party.Concentration3},
			{p.I26, party.Concentration2},
			{p.I17, party.Concentration1},
		} {
			if x.Valid {
				v := dK*(x.Float64-p.IFPlus20.Float64) - (x.PGS - party.Concentration1)
				dax[2][3+i] = strconv.FormatFloat(v, 'f', 2, 64)
			}
		}
		if p.Variation.Valid {
			v := dK * (p.I26.Float64 - p.I24.Float64)
			dax[2][8] = strconv.FormatFloat(v, 'f', 2, 64)
		}
	}

	d.SetX(left)
	d.SetFont("RobotoCondensed", "BI", 7)
	d.CellFormat(0, lineSpace1, tr("Абсолютная погрешность и вариация показаний:"),
		"", 1, "", false, 0, "")

	renderTable()

	d.SetX(left)
	sentence1("Содержание драгоценных металлов: платина ")
	sentenceB(fmt.Sprintf("%v г.", p.NobleMetalContent))
	d.Ln(lineSpace1)

	d.SetFont("RobotoCondensed-Light", "", fontSize1)
	d.SetX(left)
	d.MultiCell(width, 4, tr(fmt.Sprintf(`Ячейка соответствует комплекту документации ИБЯЛ.418425.%s и признана годной к эксплуатации. Гарантийный срок эксплуатации со дня отгрузки %d месяцев, но не более 18 месяцев со дня изготовления.`,
		p.AppliedProductTypeName, p.LifetimeMonths)),
		"", "", false)
	d.Ln(2)
	d.SetX(left)
	d.SetFont("RobotoCondensed", "B", 9)
	d.CellFormat(0, 4, tr("ОТК:"),
		"", 1, "", false, 0, "")
}
