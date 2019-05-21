package pdf

import (
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/pkg/winapp"
	"github.com/jung-kurt/gofpdf"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
)

const (
	fontSize1  = 9
	lineSpace1 = 4
)

func Run() error {

	dir, err := prepareDir()
	if err != nil {
		return err
	}
	_ = summary(dir)
	_ = passportSou(dir)
	err = passportDax(dir)
	if err != nil {
		_ = exec.Command("Explorer.exe", dir).Start()
	}
	return err
}

func newDoc() (*gofpdf.Fpdf, error) {
	fontDir, err := ensureFontDir()
	if err != nil {
		return nil, err
	}
	d := gofpdf.New("P", "mm", "A4", fontDir)
	d.AddFont("RobotoCondensed", "", "RobotoCondensed-Regular.json")
	d.AddFont("RobotoCondensed", "B", "RobotoCondensed-Bold.json")
	d.AddFont("RobotoCondensed", "I", "RobotoCondensed-Italic.json")
	d.AddFont("RobotoCondensed", "BI", "RobotoCondensed-BoldItalic.json")
	d.AddFont("RobotoCondensed-Light", "", "RobotoCondensed-Light.json")
	d.AddFont("RobotoCondensed-Light", "I", "RobotoCondensed-LightItalic.json")
	d.UnicodeTranslatorFromDescriptor("cp1251")
	d.SetLineWidth(.1)
	d.SetFillColor(225, 225, 225)
	d.SetDrawColor(169, 169, 169)
	return d, nil
}

func prepareDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", merry.WithMessage(err, "unable to locate user home catalogue")
	}
	dir := filepath.Join(usr.HomeDir, ".elco", "pdf")
	_ = os.RemoveAll(dir)
	_ = os.MkdirAll(dir, os.ModePerm)

	dir, err = ioutil.TempDir(dir, "~")
	if err != nil {
		return "", merry.WithMessage(err, "unable to create directory for pdf")
	}
	return dir, nil
}

func saveAndShowDoc(d *gofpdf.Fpdf, dir, fileName string) error {

	pdfFileName := filepath.Join(dir, fileName+".pdf")

	if err := d.OutputFileAndClose(pdfFileName); err != nil {
		return err
	}
	if err := exec.Command("explorer.exe", pdfFileName).Start(); err != nil {
		return err
	}
	return nil
}

func ensureFontDir() (string, error) {
	return winapp.ProfileFolderPath(".elco", "assets", "fonts")
}

func formatNullInt64(v sql.NullInt64) string {
	if v.Valid {
		return strconv.FormatInt(v.Int64, 10)
	}
	return ""
}

func formatNullFloat64(v sql.NullFloat64) string {
	if v.Valid {
		return formatFloat(v.Float64)
	}
	return ""
}

func formatFloat(v float64) string {
	return fmt.Sprintf("%v", v)
}

func formatNullFloat64Prec(v sql.NullFloat64, prec int) string {
	if v.Valid {
		return strconv.FormatFloat(v.Float64, 'f', prec, 64)
	}
	return ""
}
