package data

//go:generate reform

// Product represents a row in product table.
//reform:product
type Product struct {
	ProductID       int64    `reform:"product_id,pk"`
	PartyID         int64    `reform:"party_id"`
	Serial          *int64   `reform:"serial"`
	Place           int      `reform:"place"`
	ProductTypeName *string  `reform:"product_type_name"`
	Note            *string  `reform:"note"`
	IFMinus20       *float64 `reform:"i_f_minus20"`
	IFPlus20        *float64 `reform:"i_f_plus20"`
	IFPlus50        *float64 `reform:"i_f_plus50"`
	ISMinus20       *float64 `reform:"i_s_minus20"`
	ISPlus20        *float64 `reform:"i_s_plus20"`
	ISPlus50        *float64 `reform:"i_s_plus50"`
	I13             *float64 `reform:"i13"`
	I24             *float64 `reform:"i24"`
	I35             *float64 `reform:"i35"`
	I26             *float64 `reform:"i26"`
	I17             *float64 `reform:"i17"`
	NotMeasured     *float64 `reform:"not_measured"`
	Firmware        []byte   `reform:"firmware"`
	Production      bool     `reform:"production"`
	OldProductID    *string  `reform:"old_product_id"`
	OldSerial       *int64   `reform:"old_serial"`
}
