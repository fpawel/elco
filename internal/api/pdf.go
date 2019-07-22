package api

import "github.com/fpawel/elco/internal/pdf"

type PdfSvc struct {
}

func (_ PdfSvc) RunPartyID(partyID [1]int64, _ *struct{}) error {
	return pdf.RunPartyID(partyID[0])
}

func (_ PdfSvc) RunProductID(productID [1]int64, _ *struct{}) error {
	return pdf.RunProductID(productID[0])
}
