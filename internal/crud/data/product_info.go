package data

//go:generate reform

// Product represents a row in product_info table.
//reform:product_info
type ProductInfo struct {
	ProductID           int64    `reform:"product_id,pk"`
	PartyID             int64    `reform:"party_id"`
	Serial              *int64   `reform:"serial"`
	Place               int      `reform:"place"`
	ProductTypeName     *string  `reform:"product_type_name"`
	Note                *string  `reform:"note"`
	IFMinus20           *float64 `reform:"i_f_minus20"`
	IFPlus20            *float64 `reform:"i_f_plus20"`
	IFPlus50            *float64 `reform:"i_f_plus50"`
	ISMinus20           *float64 `reform:"i_s_minus20"`
	ISPlus20            *float64 `reform:"i_s_plus20"`
	ISPlus50            *float64 `reform:"i_s_plus50"`
	I13                 *float64 `reform:"i13"`
	I24                 *float64 `reform:"i24"`
	I35                 *float64 `reform:"i35"`
	I26                 *float64 `reform:"i26"`
	I17                 *float64 `reform:"i17"`
	NotMeasured         *float64 `reform:"not_measured"`
	Production          bool     `reform:"production"`
	OldProductID        *string  `reform:"old_product_id"`
	SelfProductTypeName *string  `reform:"self_product_type_name"`
	GasName             *string  `reform:"gas_name"`
	UnitsName           *string  `reform:"units_name"`
	Scale               float64  `reform:"scale"`
	NobleMetalContent   float64  `reform:"noble_metal_content"`
	LifetimeMonths      int64    `reform:"lifetime_months"`
	Lc64                bool     `reform:"lc64"`
	PointsMethod        int64    `reform:"points_method"`
	Firmware            []byte   `reform:"firmware"`
	Concentration1      float64  `reform:"concentration1"`
	Concentration3      float64  `reform:"concentration3"`
	KSens20             *float64 `reform:"k_sens20"`
	KSens50             *float64 `reform:"k_sens50"`
	DFon20              *float64 `reform:"d_fon20"`
	DFon50              *float64 `reform:"d_fon50"`
	DNotMeasured        *float64 `reform:"d_not_measured"`
	OKFon20             *bool    `reform:"ok_fon20"`
	OKDFon20            *bool    `reform:"ok_d_fon20"`
	OKKSens20           *bool    `reform:"ok_k_sens20"`
	OKDFon50            *bool    `reform:"ok_d_fon50"`
	OKKSens50           *bool    `reform:"ok_k_sens50"`
	OKDNotMeasured      *bool    `reform:"ok_d_not_measured"`
	NotOk               *bool    `reform:"not_ok"`
	HasFirmware         bool     `reform:"has_firmware"`
}
