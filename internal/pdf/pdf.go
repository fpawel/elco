package pdf

import (
	"github.com/fpawel/elco/pkg/winapp"
	"github.com/jung-kurt/gofpdf"
	"os/exec"
)

const (
	fontSize1 = 8
	lineSpace1 = 4
)

func PasportSou() error {

	fontDir, err := ensureFontDir()
	if err != nil {
		return err
	}
	d := gofpdf.New("P", "mm", "A4", fontDir)
	d.AddFont("RobotoCondensed", "", "RobotoCondensed-Regular.json")
	d.AddFont("RobotoCondensed", "B", "RobotoCondensed-Bold.json")
	d.AddPage()

	pageWidth, _ := d.GetPageSize()
	x := pageWidth/2. - 3.

	d.SetX(x)

	tr := d.UnicodeTranslatorFromDescriptor("cp1251")

	d.UnicodeTranslatorFromDescriptor("cp1251")
	d.SetLineWidth(0.1)

	d.SetFont("RobotoCondensed", "BU", 12.9)
	d.CellFormat(0, 5, tr("Электрохимическая ячейка ИБЯЛ.418425.035-100"),
		"", 1, "", false, 0, "")

	d.SetX(x)

	sentence := func(fontStyleStr string, fontSize float64, h float64, s string) {
		d.SetFont("", fontStyleStr, fontSize)
		d.CellFormat(d.GetStringWidth(tr(s)), h, tr(s), "", 0, "", false, 0, "")
	}

	sentence1 := func(fontStyleStr string, s string) {
		sentence(fontStyleStr, fontSize1 , lineSpace1 , s )
	}

	sentence1("",  "Дата изготовления: ")
	sentence1("B",  "26 февраля 2019")
	sentence1("", ".")
	sentence1("",  "Заводской номер: ")
	sentence1("B", "257")
	sentence1("",  ".")
	d.Ln(lineSpace1)
	d.SetX(x)

	sentence1("", "Фоновый ток: ")
	sentence1("B", "0.465 мкА")
	sentence1("", ".")
	sentence1("", "Чувствительность ")
	sentence1("B", "0.215 мкА/об.дол.%")
	sentence1("", ".")
	d.Ln(lineSpace1)
	d.SetX(x)

	sentence1("", "Температурная зависимость фоновых токов:")

	d.Ln(lineSpace1)
	d.SetX(x)

	table(d, []string{"T, \"C", "Iфон, мкА", "Кч, %"},
		[][]string{
			{"строка 11", "row12", "row33", },
			{"строка 21", "row22", "row33", },
			{"строка 31", "row32", "row33", },
		}, x )

	reportFileName, err := winapp.ProfileFileName(".elco", "report.pdf")
	if err != nil {
		return err
	}
	if err := d.OutputFileAndClose(reportFileName); err != nil {
		return err
	}
	if err := exec.Command("explorer.exe", reportFileName).Start(); err != nil {
		return err
	}

	return nil

}

func ensureFontDir() (string, error) {
	return winapp.ProfileFolderPath(".elco", "assets", "fonts")
}

func table(pdf *gofpdf.Fpdf, header []string, cells [][]string, left float64) {
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1251")
	// Colors, line width and bold font
	pdf.SetFillColor(225, 225, 225)
	//pdf.SetTextColor(255, 255, 255)
	pdf.SetDrawColor(169, 169, 169)
	pdf.SetLineWidth(.1)
	pdf.SetFont("", "B", fontSize1)
	// 	Header
	pdf.SetX(left)
	for _, str := range header {
		pdf.CellFormat(15, 4, tr(str), "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("", "", fontSize1)
	for _, c := range cells {
		pdf.SetX(left)
		for _, str := range c {
			pdf.CellFormat(15, 4, tr(str), "1", 0, "", false, 0, "")
		}
		pdf.Ln(-1)
	}
}
