package data

//go:generate reform

// ProductType represents a row in product_type table.
//reform:product_type
type ProductType struct {
	ProductTypeName   string  `reform:"product_type_name,pk"`
	DisplayName       *string `reform:"display_name"`
	GasName           string  `reform:"gas_name"`
	UnitsName         string  `reform:"units_name"`
	Scale             float64 `reform:"scale"`
	NobleMetalContent float64 `reform:"noble_metal_content"`
	LifetimeMonths    int64   `reform:"lifetime_months"`
	Lc64              bool    `reform:"lc64"`
	PointsMethod      int64   `reform:"points_method"`
}
