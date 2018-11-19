package data

//go:generate reform

// ProductTemperatureCurrentKSens represents a row in product_temperature_current_k_sens table.
//reform:product_temperature_current_k_sens
type ProductTemperatureCurrentKSens struct {
	ProductID   int64    `reform:"product_id"`
	Temperature float64  `reform:"temperature"`
	Current     *float64 `reform:"current"`
	KSens       *float64 `reform:"k_sens"`
}
