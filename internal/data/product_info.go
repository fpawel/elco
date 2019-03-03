package data

import (
	"database/sql"
	"time"
)

//go:generate reform

// Product represents a row in product_info table.
//reform:product_info
type ProductInfo struct {
	ProductID              int64           `reform:"product_id,pk"`
	PartyID                int64           `reform:"party_id"`
	Serial                 sql.NullInt64   `reform:"serial"`
	Place                  int             `reform:"place"`
	CreatedAt              time.Time       `reform:"created_at"`
	IFMinus20              sql.NullFloat64 `reform:"i_f_minus20"`
	IFPlus20               sql.NullFloat64 `reform:"i_f_plus20"`
	IFPlus50               sql.NullFloat64 `reform:"i_f_plus50"`
	ISMinus20              sql.NullFloat64 `reform:"i_s_minus20"`
	ISPlus20               sql.NullFloat64 `reform:"i_s_plus20"`
	ISPlus50               sql.NullFloat64 `reform:"i_s_plus50"`
	I13                    sql.NullFloat64 `reform:"i13"`
	I24                    sql.NullFloat64 `reform:"i24"`
	I35                    sql.NullFloat64 `reform:"i35"`
	I26                    sql.NullFloat64 `reform:"i26"`
	I17                    sql.NullFloat64 `reform:"i17"`
	NotMeasured            sql.NullFloat64 `reform:"not_measured"`
	KSensMinus20           sql.NullFloat64 `reform:"k_sens_minus20"`
	KSens20                sql.NullFloat64 `reform:"k_sens20"`
	KSens50                sql.NullFloat64 `reform:"k_sens50"`
	Variation              sql.NullFloat64 `reform:"variation"`
	DFon20                 sql.NullFloat64 `reform:"d_fon20"`
	DFon50                 sql.NullFloat64 `reform:"d_fon50"`
	DNotMeasured           sql.NullFloat64 `reform:"d_not_measured"`
	OKMinFon20             bool            `reform:"ok_min_fon20"`
	OKMaxFon20             bool            `reform:"ok_max_fon20"`
	OKMinFon20r            bool            `reform:"ok_min_fon20_2"`
	OKMaxFon20r            bool            `reform:"ok_max_fon20_2"`
	OKDFon20               bool            `reform:"ok_d_fon20"`
	OKMinKSens20           bool            `reform:"ok_min_k_sens20"`
	OKMaxKSens20           bool            `reform:"ok_max_k_sens20"`
	OKMinKSens50           bool            `reform:"ok_min_k_sens50"`
	OKMaxKSens50           bool            `reform:"ok_max_k_sens50"`
	OKDFon50               bool            `reform:"ok_d_fon50"`
	OKDNotMeasured         bool            `reform:"ok_d_not_measured"`
	Ok                     bool            `reform:"ok"`
	HasFirmware            bool            `reform:"has_firmware"`
	Production             bool            `reform:"production"`
	AppliedProductTypeName string          `reform:"applied_product_type_name"`
	GasCode                byte            `reform:"gas_code"`
	UnitsCode              byte            `reform:"units_code"`
	GasName                string          `reform:"gas_name"`
	UnitsName              string          `reform:"units_name"`
	Scale                  float64         `reform:"scale"`
	NobleMetalContent      float64         `reform:"noble_metal_content"`
	LifetimeMonths         int64           `reform:"lifetime_months"`
	PointsMethod           sql.NullInt64   `reform:"points_method"`
	AppliedPointsMethod    int64           `reform:"applied_points_method"`
	ProductTypeName        sql.NullString  `reform:"product_type_name"`
	Note                   sql.NullString  `reform:"note"`
}
