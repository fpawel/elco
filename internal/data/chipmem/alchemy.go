package chipmem

import (
	"database/sql"
	"github.com/fpawel/elco/internal/data"
)

func CalculateFonMinus20() error {

	for _, p := range data.ProductsInfoAll(data.LastPartyID()) {
		t, err := ProductInfo{p}.TableFon()
		if err != nil {
			continue
		}
		if err := data.SetProductValue(p.ProductID, "i_f_minus20", data.NewApproximationTable(t).F(-20)); err != nil {
			return err
		}
	}
	return nil
}

func CalculateSensMinus20(k float64) error {
	for _, p := range data.ProductsInfoAll(data.LastPartyID()) {
		if !(p.IFPlus20.Valid && p.ISPlus20.Valid) {
			continue
		}

		if !p.IFMinus20.Valid {
			p.IFMinus20 = sql.NullFloat64{
				Float64: data.NewApproximationTable(data.TableXY{
					20: p.IFPlus20.Float64,
				}).F(-20),
				Valid: true,
			}

			if err := data.SetProductValue(p.ProductID, "i_f_minus20", p.IFMinus20.Float64); err != nil {
				return err
			}
		}

		ISMinus20 :=
			p.IFMinus20.Float64 + (p.ISPlus20.Float64-p.IFPlus20.Float64)*k/100.
		if err := data.SetProductValue(p.ProductID, "i_s_minus20", ISMinus20); err != nil {
			return err
		}
	}
	return nil
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
