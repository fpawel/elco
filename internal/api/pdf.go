package api

import "github.com/fpawel/elco/internal/pdf"

type PdfSvc struct {
}

func (_ PdfSvc) Run(partyID [1]int64, _ *struct{}) error {
	return pdf.Run(partyID[0])
}
