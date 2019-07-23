package pdf

import (
	"fmt"
	"github.com/fpawel/elco/internal/data"
	"github.com/jung-kurt/gofpdf"
	"math"
	"sort"
	"strconv"
)

func passportTempCodes(dir string, party data.Party) error {

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
		doPassportTempCodes(d, spaceX, width, party.Products[i])
		d.SetY(y)
		if i == len(party.Products)-1 {
			break
		}
		doPassportTempCodes(d, pageWidth/2., width, party.Products[i+1])

	}

	if err := saveAndShowDoc(d, dir, "temp_codes"); err != nil {
		return err
	}

	return nil
}

func doPassportTempCodes(d *gofpdf.Fpdf, left, width float64, p data.ProductInfo) {

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
	code1, code2 := tempCodes(p)
	sentence1("Коды температурных характеристик: №1=")
	sentenceB(code1)
	sentence1(", №2=")
	sentenceB(code2)
	sentence1(".")
	d.Ln(lineSpace1)

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

func tempCodes(p data.ProductInfo) (code1, code2 string) {
	if p.DFon50.Valid {
		code1 = strconv.Itoa(tempCode1(p.DFon50.Float64))
	}
	if p.KSens50.Valid && p.KSensMinus20.Valid {
		code2 = strconv.Itoa(tempCode2(p.KSensMinus20.Float64, p.KSens50.Float64))
	}
	return
}

func tempCode1(dFon50 float64) int {

	type T = struct {
		n int
		a float64
	}
	var xs []T
	for i, v := range []float64{0, 0.3, 0.6, 0.9, 1.2, 1.5, -0.3, -0.6, -0.9, -1.2, -1.5} {
		xs = append(xs, T{i + 1, v})
	}
	sort.Slice(xs, func(i, j int) bool {
		return math.Abs(dFon50-xs[i].a) < math.Abs(dFon50-xs[j].a)
	})
	return xs[0].n
}

func tempCode2(k20, k50 float64) int {

	type T struct {
		n   int
		k50 float64
	}
	var xs []T

	for i, v := range []float64{100, 110, 120, 130} {
		xs = append(xs, T{i, v})
	}
	sort.Slice(xs, func(i, j int) bool {
		return math.Abs(xs[i].k50-k50) < math.Abs(xs[j].k50-k50)
	})
	n := xs[0].n*2 + 1

	if math.Abs(k20-60) < math.Abs(k20-40) {
		n++
	}
	return n
}
