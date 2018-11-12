package db

//go:generate reform

// ProductType represents a row in product_type table.
//reform:product_type
type ProductType struct {
	ProductTypeName   string   `reform:"product_type_name,pk"`
	DisplayName       *string  `reform:"display_name"`
	GasName           string   `reform:"gas_name"`
	UnitsName         string   `reform:"units_name"`
	Scale             float64  `reform:"scale"`
	NobleMetalContent float64  `reform:"noble_metal_content"`
	LifetimeMonths    int64    `reform:"lifetime_months"`
	Lc64              bool     `reform:"lc64"`
	PointsMethod      int64    `reform:"points_method"`
	MaxFon            *float64 `reform:"max_fon"`
	MaxDFon           *float64 `reform:"max_d_fon"`
	MinKSens20        *float64 `reform:"min_k_sens20"`
	MaxKSens20        *float64 `reform:"max_k_sens20"`
	MinDTemp          *float64 `reform:"min_d_temp"`
	MaxDTemp          *float64 `reform:"max_d_temp"`
	MinKSens50        *float64 `reform:"min_k_sens50"`
	MaxKSens50        *float64 `reform:"max_k_sens50"`
	MaxDNotMeasured   *float64 `reform:"max_d_not_measured"`
}
