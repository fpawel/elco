package data

import "database/sql"

//go:generate reform

// Product represents a row in product_info table.
//reform:product_info
type ProductInfo struct {
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
	Production      bool            `reform:"production"`
	KSens20         sql.NullFloat64 `reform:"k_sens20"`
	KSens50         sql.NullFloat64 `reform:"k_sens50"`
	DFon20          sql.NullFloat64 `reform:"d_fon20"`
	DFon50          sql.NullFloat64 `reform:"d_fon50"`
	DNotMeasured    sql.NullFloat64 `reform:"d_not_measured"`
	OKFon20         sql.NullBool    `reform:"ok_fon20"`
	OKDFon20        sql.NullBool    `reform:"ok_d_fon20"`
	OKKSens20       sql.NullBool    `reform:"ok_k_sens20"`
	OKDFon50        sql.NullBool    `reform:"ok_d_fon50"`
	OKKSens50       sql.NullBool    `reform:"ok_k_sens50"`
	OKDNotMeasured  sql.NullBool    `reform:"ok_d_not_measured"`
	NotOk           sql.NullBool    `reform:"not_ok"`
	HasFirmware     bool            `reform:"has_firmware"`
}
