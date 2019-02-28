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

	x, y := d.GetXY()

	d.SetXY(x+100, y)

	tr := d.UnicodeTranslatorFromDescriptor("cp1251")

	d.UnicodeTranslatorFromDescriptor("cp1251")
	d.SetLineWidth(0.3)

	d.SetFont("RobotoCondensed", "BU", 12)
	d.CellFormat(160, 5, tr("Электрохимическая ячейка ИБЯЛ.418425.035-100"),
		"", 1, "", false, 0, "")

	d.SetFont("RobotoCondensed", "", 10)
	d.CellFormat(160, 5, tr("Дата изготовления: 26 февраля 2019. Заводской номер: 257."),
		"", 1, "", false, 0, "")
	d.CellFormat(160, 5, tr("Фоновый ток: 0.465 мкА. Чувствительность 0.215 мкА/об.дол.%."),
		"", 1, "", false, 0, "")

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
