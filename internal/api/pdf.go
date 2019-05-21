package api

import "github.com/fpawel/elco/internal/pdf"

type PdfSvc struct {
}

func (_ PdfSvc) Run(_ struct{}, _ *struct{}) error {
	return pdf.Run()
}
