package api

import (
	"fmt"
	"github.com/fpawel/elco/internal/data"
	"strconv"
)

type PartiesCatalogueSvc struct {
}

func (_ *PartiesCatalogueSvc) YearsMonths(_ struct{}, r *[]YearMonth) error {
	if err := data.DBx.Select(r, `SELECT DISTINCT year, month FROM party_info ORDER BY year DESC, month DESC`); err != nil {
		panic(err)
	}
	return nil
}

func (_ *PartiesCatalogueSvc) PartiesOfYearMonth(x YearMonth, r *[]Party2) error {
	if err := data.DBx.Select(r, `
SELECT cast(strftime('%d', DATETIME(created_at, '+3 hours')) AS INTEGER)  AS day, 
       party_id, note, product_type_name,
       party_id = (SELECT party_id FROM party ORDER BY created_at DESC LIMIT 1)  AS last
FROM party
WHERE cast(strftime('%Y', DATETIME(created_at, '+3 hours')) AS INTEGER) = ?
  AND cast(strftime('%m', DATETIME(created_at, '+3 hours')) AS INTEGER) = ?
ORDER BY created_at`, x.Year, x.Month); err != nil {
		panic(err)
	}
	return nil
}

func (_ *PartiesCatalogueSvc) Party(a [1]int64, r *Party1) error {
	*r = newParty1(a[0])
	return nil
}

func (x *PartiesCatalogueSvc) DeletePartyID(partyID [1]int64, _ *struct{}) error {
	if _, err := data.DB.Exec(`DELETE FROM party WHERE party_id = ?`, partyID[0]); err != nil {
		return err
	}
	return nil
}

func (x *PartiesCatalogueSvc) SetProductsProduction(v struct {
	ProductIDs []int64
	Production bool
}, _ *struct{}) error {
	var s string
	for i, productID := range v.ProductIDs {
		if i > 0 {
			s += ","
		}
		s += strconv.FormatInt(productID, 10)
	}
	query := fmt.Sprintf(
		"UPDATE product SET production = ? WHERE product_id IN (%s)", s)

	data.DBx.MustExec(query, v.Production)

	return nil
}

func (x *PartiesCatalogueSvc) ToggleProductProduction(productID [1]int64, _ *struct{}) error {
	var p data.Product
	if err := data.DB.FindByPrimaryKeyTo(&p, productID[0]); err != nil {
		return err
	}
	p.Production = !p.Production
	if err := data.DB.Save(&p); err != nil {
		return err
	}
	return nil
}
