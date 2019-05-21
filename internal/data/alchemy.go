package data

func CalculateFonMinus20() error {

	party := GetLastPartyWithProductsInfo(ProductsFilter{})
	for _, p := range party.Products {
		t, err := p.TableFon()
		if err != nil {
			continue
		}
		if err := SetProductValue(p.ProductID, "i_f_minus20", NewApproximationTable(t).F(-20)); err != nil {
			return err
		}
	}
	return nil
}

func CalculateSensMinus20(k float64) error {
	party := GetLastPartyWithProductsInfo(ProductsFilter{})
	for _, p := range party.Products {
		if !(p.IFPlus20.Valid && p.ISPlus20.Valid && p.IFMinus20.Valid) {
			continue
		}
		ISMinus20 :=
			p.IFMinus20.Float64 + (p.ISPlus20.Float64-p.IFPlus20.Float64)*k/100.
		if err := SetProductValue(p.ProductID, "i_s_minus20", ISMinus20); err != nil {
			return err
		}
	}
	return nil
}

func CalculateSensPlus50(k float64) error {
	party := GetLastPartyWithProductsInfo(ProductsFilter{})
	for _, p := range party.Products {
		if !(p.IFPlus20.Valid && p.ISPlus20.Valid && p.IFPlus50.Valid) {
			continue
		}
		ISPlus50 :=
			p.IFPlus50.Float64 + (p.ISPlus20.Float64-p.IFPlus20.Float64)*k/100.

		if err := SetProductValue(p.ProductID, "i_s_plus50", ISPlus50); err != nil {
			return err
		}
	}
	return nil
}
