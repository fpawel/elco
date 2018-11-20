package data

//go:generate reform

// Gas represents a row in gas table.
//reform:gas
type Gas struct {
	GasName string `reform:"gas_name,pk"`
	Code    byte   `reform:"code"`
}
