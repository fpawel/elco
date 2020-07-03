package chipmem

import (
	"github.com/fpawel/elco/internal/data"
)

func CalculateMinus20(k float64) error {
	for _, p := range data.ProductsInfoAll(data.LastPartyID()) {
		fon, sens := calculateMinus20Product(p, k)
		_, err := data.DBx.Exec(`UPDATE product SET i_f_minus20 = ?, i_s_minus20 = ? WHERE product_id = ?`,
			fon, sens, p.ProductID)
		if err != nil {
			return err
		}
	}
	return nil
}

type nullFloat = *float64

func floatVal(x float64) nullFloat {
	return &x
}

var noVal nullFloat

func calculateMinus20Product(p data.ProductInfo, k float64) (nullFloat, nullFloat) {
	if !(p.IFPlus20.Valid && p.ISPlus20.Valid) {
		return noVal, noVal
	}
	t, err := ProductInfo{p}.TableFon2()
	if err != nil {
		return noVal, noVal
	}
	f := data.NewApproximationTable(t).F(-20)
	return floatVal(f), floatVal(f + (p.ISPlus20.Float64-p.IFPlus20.Float64)*k/100.)
}

func CalculateSensPlus50(k float64) error {
	for _, p := range data.ProductsInfoAll(data.LastPartyID()) {
		if !(p.IFPlus20.Valid && p.ISPlus20.Valid && p.IFPlus50.Valid) {
			continue
		}
		ISPlus50 :=
			p.IFPlus50.Float64 + (p.ISPlus20.Float64-p.IFPlus20.Float64)*k/100.

		if err := data.SetProductValue(p.ProductID, "i_s_plus50", ISPlus50); err != nil {
			return err
		}
	}
	return nil
}
