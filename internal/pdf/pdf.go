package pdf

import (
	"github.com/fpawel/elco/pkg/winapp"
	"github.com/jung-kurt/gofpdf"
	"os/exec"
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

	sentence := func(fontStyleStr string, fontSize float64, h int, s string) {
		d.SetFont("", fontStyleStr, fontSize)
		d.CellFormat(d.GetStringWidth(tr(s)), 5, tr(s), "", 0, "", false, 0, "")
	}
	sentence("", 10, 5, "Дата изготовления ")
	sentence("B", 10, 5, "26 февраля 2019 ")
	sentence("", 10, 5, "Заводской номер ")
	sentence("B", 10, 5, "257")
	d.Ln(5)
	d.SetX(x)
	sentence("", 10, 5, "Фоновый ток ")
	sentence("B", 10, 5, "0.465 мкА ")
	sentence("", 10, 5, "Чувствительность ")
	sentence("B", 10, 5, "0.215 мкА/об.дол.%")
	d.Ln(5)
	d.SetX(x)

	fancyTable(d, []string{"tab1", "tab2", "tab3", "tab4"},
		[][]string{
			{"row11", "row12", "row33", "row44"},
			{"row21", "row22", "row33", "row44"},
			{"row31", "row32", "row33", "row44"},
			{"row41", "row42", "row33", "row44"},
		}, x)

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

func fancyTable(pdf *gofpdf.Fpdf, header []string, cells [][]string, left float64) {
	// Colors, line width and bold font
	pdf.SetFillColor(230, 230, 230)
	//pdf.SetTextColor(255, 255, 255)
	pdf.SetDrawColor(169, 169, 169)
	pdf.SetLineWidth(.1)
	pdf.SetFont("", "B", 8)
	// 	Header
	pdf.SetX(left)
	for _, str := range header {
		pdf.CellFormat(15, 4, str, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
	// Color and font restoration
	pdf.SetFillColor(240, 240, 240)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("", "", 8)
	// 	Data
	fill := false
	for _, c := range cells {
		pdf.SetX(left)
		for _, cellStr := range c {
			pdf.CellFormat(15, 4, cellStr, "1", 0, "", fill, 0, "")
		}
		pdf.Ln(-1)
		fill = !fill

	}
}
