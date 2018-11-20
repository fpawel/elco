package data

//go:generate reform

// Units represents a row in units table.
//reform:units
type Units struct {
	UnitsName string `reform:"units_name,pk"`
	Code      byte   `reform:"code"`
}
