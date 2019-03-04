package data

import (
	"gopkg.in/reform.v1"
)

func CalculateFonMinus20(db *reform.DB, party *Party) error {
	if err := GetLastPartyWithProductsInfo(db, ProductsFilter{}, party); err != nil {
		return err
	}
	for i, p := range party.Products{
		t, err := p.TableFon()
		if err != nil {
			continue
		}
		a := NewApproximationTable(t)
		var product Product
		if err := db.FindByPrimaryKeyTo(&product, p.ProductID); err != nil {
			return err
		}
		product.IFMinus20.Valid = true
		product.IFMinus20.Float64 = a.F(-20)
		if err := db.Save(&product); err != nil {
			return err
		}
		party.Products[i].IFMinus20 = product.IFMinus20
	}
	return nil
}

func CalculateSensMinus20(db *reform.DB, k float64, party *Party) error {
	if err := GetLastPartyWithProductsInfo(db, ProductsFilter{}, party); err != nil {
		return err
	}
	for i, p := range party.Products{
		if !(p.IFPlus20.Valid && p.ISPlus20.Valid && p.IFMinus20.Valid) {
			continue
		}
		var product Product
		if err := db.FindByPrimaryKeyTo(&product, p.ProductID); err != nil {
			return err
		}
		product.ISMinus20.Valid = true
		product.ISMinus20.Float64 =
			product.IFMinus20.Float64 + (product.ISPlus20.Float64 - product.IFPlus20.Float64) * k / 100.
		if err := db.Save(&product); err != nil {
			return err
		}
		party.Products[i].ISMinus20 = product.ISMinus20
	}
	return nil
}

func CalculateSensPlus50(db *reform.DB, k float64, party *Party) error {
	if err := GetLastPartyWithProductsInfo(db, ProductsFilter{}, party); err != nil {
		return err
	}
	for i, p := range party.Products{
		if !(p.IFPlus20.Valid && p.ISPlus20.Valid && p.IFPlus50.Valid) {
			continue
		}
		var product Product
		if err := db.FindByPrimaryKeyTo(&product, p.ProductID); err != nil {
			return err
		}
		product.ISPlus50.Valid = true
		product.ISPlus50.Float64 =
			product.IFPlus50.Float64 + (product.ISPlus20.Float64 - product.IFPlus20.Float64) * k / 100.
		if err := db.Save(&product); err != nil {
			return err
		}
		party.Products[i].ISPlus50 = product.ISPlus50
	}
	return nil
}
