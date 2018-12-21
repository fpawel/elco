package data

import "database/sql"

//go:generate reform

// Product represents a row in product table.
//reform:product
type Product struct {
	ProductID       int64           `reform:"product_id,pk"`
	PartyID         int64           `reform:"party_id"`
	Serial          sql.NullInt64   `reform:"serial"`
	Place           int             `reform:"place"`
	ProductTypeName sql.NullString  `reform:"product_type_name"`
	Note            sql.NullString  `reform:"note"`
	IFMinus20       sql.NullFloat64 `reform:"i_f_minus20"`
	IFPlus20        sql.NullFloat64 `reform:"i_f_plus20"`
	IFPlus50        sql.NullFloat64 `reform:"i_f_plus50"`
	ISMinus20       sql.NullFloat64 `reform:"i_s_minus20"`
	ISPlus20        sql.NullFloat64 `reform:"i_s_plus20"`
	ISPlus50        sql.NullFloat64 `reform:"i_s_plus50"`
	I13             sql.NullFloat64 `reform:"i13"`
	I24             sql.NullFloat64 `reform:"i24"`
	I35             sql.NullFloat64 `reform:"i35"`
	I26             sql.NullFloat64 `reform:"i26"`
	I17             sql.NullFloat64 `reform:"i17"`
	NotMeasured     sql.NullFloat64 `reform:"not_measured"`
	Firmware        []byte          `reform:"firmware"`
	Production      bool            `reform:"production"`
	OldProductID    sql.NullString  `reform:"old_product_id"`
	OldSerial       sql.NullInt64   `reform:"old_serial"`
	PointsMethod    sql.NullInt64   `reform:"points_method"`
}
