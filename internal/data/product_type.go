package data

//go:generate reform

// ProductType represents a row in product_type table.
//reform:product_type
type ProductType struct {
	ProductTypeName   string  `reform:"product_type_name,pk"`
	GasName           string  `reform:"gas_name"`
	UnitsName         string  `reform:"units_name"`
	Scale             float64 `reform:"scale"`
	NobleMetalContent float64 `reform:"noble_metal_content"`
	LifetimeMonths    int64   `reform:"lifetime_months"`
}
