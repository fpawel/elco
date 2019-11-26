package data

import "database/sql"

//go:generate reform

// ProductType represents a row in product_type table.
//reform:product_type
type ProductType struct {
	ProductTypeName   string          `reform:"product_type_name,pk"`
	GasName           string          `reform:"gas_name"`
	UnitsName         string          `reform:"units_name"`
	Scale             float64         `reform:"scale"`
	NobleMetalContent float64         `reform:"noble_metal_content"`
	LifetimeMonths    int64           `reform:"lifetime_months"`
	MinFon            sql.NullFloat64 `reform:"min_fon"`
	MaxFon            sql.NullFloat64 `reform:"max_fon"`
	MaxDFon           sql.NullFloat64 `reform:"max_d_fon"`
	MinKSens20        sql.NullFloat64 `reform:"min_k_sens20"`
	MaxKSens20        sql.NullFloat64 `reform:"max_k_sens20"`
	MinKSens50        sql.NullFloat64 `reform:"min_k_sens50"`
	MaxKSens50        sql.NullFloat64 `reform:"max_k_sens50"`
	MinDTemp          sql.NullFloat64 `reform:"min_d_temp"`
	MaxDTemp          sql.NullFloat64 `reform:"max_d_temp"`
	MaxDNotMeasured   sql.NullFloat64 `reform:"max_d_not_measured"`
	KSens20           sql.NullFloat64 `reform:"k_sens20"`
	Fon20             sql.NullFloat64 `reform:"fon20"`
	MaxD1             sql.NullFloat64 `reform:"max_d1"`
	MaxD2             sql.NullFloat64 `reform:"max_d2"`
	MaxD3             sql.NullFloat64 `reform:"max_d3"`
	PointsMethod      int64           `reform:"points_method"`
}
