// Code generated by gopkg.in/reform.v1. DO NOT EDIT.

package data

import (
	"fmt"
	"strings"

	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/parse"
)

type productInfoTableType struct {
	s parse.StructInfo
	z []interface{}
}

// Schema returns a schema name in SQL database ("").
func (v *productInfoTableType) Schema() string {
	return v.s.SQLSchema
}

// Name returns a view or table name in SQL database ("product_info").
func (v *productInfoTableType) Name() string {
	return v.s.SQLName
}

// Columns returns a new slice of column names for that view or table in SQL database.
func (v *productInfoTableType) Columns() []string {
	return []string{"product_id", "party_id", "serial", "place", "created_at", "i_f_minus20", "i_f_plus20", "i_f_plus50", "i_s_minus20", "i_s_plus20", "i_s_plus50", "i13", "i24", "i35", "i26", "i17", "not_measured", "k_sens_minus20", "k_sens20", "k_sens50", "d_fon20", "d_fon50", "d_not_measured", "ok_fon20", "ok_d_fon20", "ok_k_sens20", "ok_d_fon50", "ok_k_sens50", "ok_d_not_measured", "not_ok", "has_firmware", "production", "applied_product_type_name", "gas_code", "units_code", "gas_name", "units_name", "scale", "noble_metal_content", "lifetime_months", "lc64", "points_method", "product_type_name", "note"}
}

// NewStruct makes a new struct for that view or table.
func (v *productInfoTableType) NewStruct() reform.Struct {
	return new(ProductInfo)
}

// NewRecord makes a new record for that table.
func (v *productInfoTableType) NewRecord() reform.Record {
	return new(ProductInfo)
}

// PKColumnIndex returns an index of primary key column for that table in SQL database.
func (v *productInfoTableType) PKColumnIndex() uint {
	return uint(v.s.PKFieldIndex)
}

// ProductInfoTable represents product_info view or table in SQL database.
var ProductInfoTable = &productInfoTableType{
	s: parse.StructInfo{Type: "ProductInfo", SQLSchema: "", SQLName: "product_info", Fields: []parse.FieldInfo{{Name: "ProductID", Type: "int64", Column: "product_id"}, {Name: "PartyID", Type: "int64", Column: "party_id"}, {Name: "Serial", Type: "sql.NullInt64", Column: "serial"}, {Name: "Place", Type: "int", Column: "place"}, {Name: "CreatedAt", Type: "time.Time", Column: "created_at"}, {Name: "IFMinus20", Type: "sql.NullFloat64", Column: "i_f_minus20"}, {Name: "IFPlus20", Type: "sql.NullFloat64", Column: "i_f_plus20"}, {Name: "IFPlus50", Type: "sql.NullFloat64", Column: "i_f_plus50"}, {Name: "ISMinus20", Type: "sql.NullFloat64", Column: "i_s_minus20"}, {Name: "ISPlus20", Type: "sql.NullFloat64", Column: "i_s_plus20"}, {Name: "ISPlus50", Type: "sql.NullFloat64", Column: "i_s_plus50"}, {Name: "I13", Type: "sql.NullFloat64", Column: "i13"}, {Name: "I24", Type: "sql.NullFloat64", Column: "i24"}, {Name: "I35", Type: "sql.NullFloat64", Column: "i35"}, {Name: "I26", Type: "sql.NullFloat64", Column: "i26"}, {Name: "I17", Type: "sql.NullFloat64", Column: "i17"}, {Name: "NotMeasured", Type: "sql.NullFloat64", Column: "not_measured"}, {Name: "KSensMinus20", Type: "sql.NullFloat64", Column: "k_sens_minus20"}, {Name: "KSens20", Type: "sql.NullFloat64", Column: "k_sens20"}, {Name: "KSens50", Type: "sql.NullFloat64", Column: "k_sens50"}, {Name: "DFon20", Type: "sql.NullFloat64", Column: "d_fon20"}, {Name: "DFon50", Type: "sql.NullFloat64", Column: "d_fon50"}, {Name: "DNotMeasured", Type: "sql.NullFloat64", Column: "d_not_measured"}, {Name: "OKFon20", Type: "sql.NullBool", Column: "ok_fon20"}, {Name: "OKDFon20", Type: "sql.NullBool", Column: "ok_d_fon20"}, {Name: "OKKSens20", Type: "sql.NullBool", Column: "ok_k_sens20"}, {Name: "OKDFon50", Type: "sql.NullBool", Column: "ok_d_fon50"}, {Name: "OKKSens50", Type: "sql.NullBool", Column: "ok_k_sens50"}, {Name: "OKDNotMeasured", Type: "sql.NullBool", Column: "ok_d_not_measured"}, {Name: "NotOk", Type: "sql.NullBool", Column: "not_ok"}, {Name: "HasFirmware", Type: "bool", Column: "has_firmware"}, {Name: "Production", Type: "bool", Column: "production"}, {Name: "AppliedProductTypeName", Type: "string", Column: "applied_product_type_name"}, {Name: "GasCode", Type: "uint8", Column: "gas_code"}, {Name: "UnitsCode", Type: "uint8", Column: "units_code"}, {Name: "GasName", Type: "string", Column: "gas_name"}, {Name: "UnitsName", Type: "string", Column: "units_name"}, {Name: "Scale", Type: "float64", Column: "scale"}, {Name: "NobleMetalContent", Type: "float64", Column: "noble_metal_content"}, {Name: "LifetimeMonths", Type: "int64", Column: "lifetime_months"}, {Name: "Lc64", Type: "bool", Column: "lc64"}, {Name: "PointsMethod", Type: "int64", Column: "points_method"}, {Name: "ProductTypeName", Type: "sql.NullString", Column: "product_type_name"}, {Name: "Note", Type: "sql.NullString", Column: "note"}}, PKFieldIndex: 0},
	z: new(ProductInfo).Values(),
}

// String returns a string representation of this struct or record.
func (s ProductInfo) String() string {
	res := make([]string, 44)
	res[0] = "ProductID: " + reform.Inspect(s.ProductID, true)
	res[1] = "PartyID: " + reform.Inspect(s.PartyID, true)
	res[2] = "Serial: " + reform.Inspect(s.Serial, true)
	res[3] = "Place: " + reform.Inspect(s.Place, true)
	res[4] = "CreatedAt: " + reform.Inspect(s.CreatedAt, true)
	res[5] = "IFMinus20: " + reform.Inspect(s.IFMinus20, true)
	res[6] = "IFPlus20: " + reform.Inspect(s.IFPlus20, true)
	res[7] = "IFPlus50: " + reform.Inspect(s.IFPlus50, true)
	res[8] = "ISMinus20: " + reform.Inspect(s.ISMinus20, true)
	res[9] = "ISPlus20: " + reform.Inspect(s.ISPlus20, true)
	res[10] = "ISPlus50: " + reform.Inspect(s.ISPlus50, true)
	res[11] = "I13: " + reform.Inspect(s.I13, true)
	res[12] = "I24: " + reform.Inspect(s.I24, true)
	res[13] = "I35: " + reform.Inspect(s.I35, true)
	res[14] = "I26: " + reform.Inspect(s.I26, true)
	res[15] = "I17: " + reform.Inspect(s.I17, true)
	res[16] = "NotMeasured: " + reform.Inspect(s.NotMeasured, true)
	res[17] = "KSensMinus20: " + reform.Inspect(s.KSensMinus20, true)
	res[18] = "KSens20: " + reform.Inspect(s.KSens20, true)
	res[19] = "KSens50: " + reform.Inspect(s.KSens50, true)
	res[20] = "DFon20: " + reform.Inspect(s.DFon20, true)
	res[21] = "DFon50: " + reform.Inspect(s.DFon50, true)
	res[22] = "DNotMeasured: " + reform.Inspect(s.DNotMeasured, true)
	res[23] = "OKFon20: " + reform.Inspect(s.OKFon20, true)
	res[24] = "OKDFon20: " + reform.Inspect(s.OKDFon20, true)
	res[25] = "OKKSens20: " + reform.Inspect(s.OKKSens20, true)
	res[26] = "OKDFon50: " + reform.Inspect(s.OKDFon50, true)
	res[27] = "OKKSens50: " + reform.Inspect(s.OKKSens50, true)
	res[28] = "OKDNotMeasured: " + reform.Inspect(s.OKDNotMeasured, true)
	res[29] = "NotOk: " + reform.Inspect(s.NotOk, true)
	res[30] = "HasFirmware: " + reform.Inspect(s.HasFirmware, true)
	res[31] = "Production: " + reform.Inspect(s.Production, true)
	res[32] = "AppliedProductTypeName: " + reform.Inspect(s.AppliedProductTypeName, true)
	res[33] = "GasCode: " + reform.Inspect(s.GasCode, true)
	res[34] = "UnitsCode: " + reform.Inspect(s.UnitsCode, true)
	res[35] = "GasName: " + reform.Inspect(s.GasName, true)
	res[36] = "UnitsName: " + reform.Inspect(s.UnitsName, true)
	res[37] = "Scale: " + reform.Inspect(s.Scale, true)
	res[38] = "NobleMetalContent: " + reform.Inspect(s.NobleMetalContent, true)
	res[39] = "LifetimeMonths: " + reform.Inspect(s.LifetimeMonths, true)
	res[40] = "Lc64: " + reform.Inspect(s.Lc64, true)
	res[41] = "PointsMethod: " + reform.Inspect(s.PointsMethod, true)
	res[42] = "ProductTypeName: " + reform.Inspect(s.ProductTypeName, true)
	res[43] = "Note: " + reform.Inspect(s.Note, true)
	return strings.Join(res, ", ")
}

// Values returns a slice of struct or record field values.
// Returned interface{} values are never untyped nils.
func (s *ProductInfo) Values() []interface{} {
	return []interface{}{
		s.ProductID,
		s.PartyID,
		s.Serial,
		s.Place,
		s.CreatedAt,
		s.IFMinus20,
		s.IFPlus20,
		s.IFPlus50,
		s.ISMinus20,
		s.ISPlus20,
		s.ISPlus50,
		s.I13,
		s.I24,
		s.I35,
		s.I26,
		s.I17,
		s.NotMeasured,
		s.KSensMinus20,
		s.KSens20,
		s.KSens50,
		s.DFon20,
		s.DFon50,
		s.DNotMeasured,
		s.OKFon20,
		s.OKDFon20,
		s.OKKSens20,
		s.OKDFon50,
		s.OKKSens50,
		s.OKDNotMeasured,
		s.NotOk,
		s.HasFirmware,
		s.Production,
		s.AppliedProductTypeName,
		s.GasCode,
		s.UnitsCode,
		s.GasName,
		s.UnitsName,
		s.Scale,
		s.NobleMetalContent,
		s.LifetimeMonths,
		s.Lc64,
		s.PointsMethod,
		s.ProductTypeName,
		s.Note,
	}
}

// Pointers returns a slice of pointers to struct or record fields.
// Returned interface{} values are never untyped nils.
func (s *ProductInfo) Pointers() []interface{} {
	return []interface{}{
		&s.ProductID,
		&s.PartyID,
		&s.Serial,
		&s.Place,
		&s.CreatedAt,
		&s.IFMinus20,
		&s.IFPlus20,
		&s.IFPlus50,
		&s.ISMinus20,
		&s.ISPlus20,
		&s.ISPlus50,
		&s.I13,
		&s.I24,
		&s.I35,
		&s.I26,
		&s.I17,
		&s.NotMeasured,
		&s.KSensMinus20,
		&s.KSens20,
		&s.KSens50,
		&s.DFon20,
		&s.DFon50,
		&s.DNotMeasured,
		&s.OKFon20,
		&s.OKDFon20,
		&s.OKKSens20,
		&s.OKDFon50,
		&s.OKKSens50,
		&s.OKDNotMeasured,
		&s.NotOk,
		&s.HasFirmware,
		&s.Production,
		&s.AppliedProductTypeName,
		&s.GasCode,
		&s.UnitsCode,
		&s.GasName,
		&s.UnitsName,
		&s.Scale,
		&s.NobleMetalContent,
		&s.LifetimeMonths,
		&s.Lc64,
		&s.PointsMethod,
		&s.ProductTypeName,
		&s.Note,
	}
}

// View returns View object for that struct.
func (s *ProductInfo) View() reform.View {
	return ProductInfoTable
}

// Table returns Table object for that record.
func (s *ProductInfo) Table() reform.Table {
	return ProductInfoTable
}

// PKValue returns a value of primary key for that record.
// Returned interface{} value is never untyped nil.
func (s *ProductInfo) PKValue() interface{} {
	return s.ProductID
}

// PKPointer returns a pointer to primary key field for that record.
// Returned interface{} value is never untyped nil.
func (s *ProductInfo) PKPointer() interface{} {
	return &s.ProductID
}

// HasPK returns true if record has non-zero primary key set, false otherwise.
func (s *ProductInfo) HasPK() bool {
	return s.ProductID != ProductInfoTable.z[ProductInfoTable.s.PKFieldIndex]
}

// SetPK sets record primary key.
func (s *ProductInfo) SetPK(pk interface{}) {
	if i64, ok := pk.(int64); ok {
		s.ProductID = int64(i64)
	} else {
		s.ProductID = pk.(int64)
	}
}

// check interfaces
var (
	_ reform.View   = ProductInfoTable
	_ reform.Struct = (*ProductInfo)(nil)
	_ reform.Table  = ProductInfoTable
	_ reform.Record = (*ProductInfo)(nil)
	_ fmt.Stringer  = (*ProductInfo)(nil)
)

func init() {
	parse.AssertUpToDate(&ProductInfoTable.s, new(ProductInfo))
}
