// Code generated by gopkg.in/reform.v1. DO NOT EDIT.

package data

import (
	"fmt"
	"strings"

	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/parse"
)

type productTableType struct {
	s parse.StructInfo
	z []interface{}
}

// Schema returns a schema name in SQL database ("").
func (v *productTableType) Schema() string {
	return v.s.SQLSchema
}

// Name returns a view or table name in SQL database ("product").
func (v *productTableType) Name() string {
	return v.s.SQLName
}

// Columns returns a new slice of column names for that view or table in SQL database.
func (v *productTableType) Columns() []string {
	return []string{"product_id", "party_id", "serial", "place", "product_type_name", "note", "i_f_minus20", "i_f_plus20", "i_f_plus50", "i_s_minus20", "i_s_plus20", "i_s_plus50", "i13", "i24", "i35", "i26", "i17", "not_measured", "firmware", "production", "old_product_id", "old_serial", "points_method"}
}

// NewStruct makes a new struct for that view or table.
func (v *productTableType) NewStruct() reform.Struct {
	return new(Product)
}

// NewRecord makes a new record for that table.
func (v *productTableType) NewRecord() reform.Record {
	return new(Product)
}

// PKColumnIndex returns an index of primary key column for that table in SQL database.
func (v *productTableType) PKColumnIndex() uint {
	return uint(v.s.PKFieldIndex)
}

// ProductTable represents product view or table in SQL database.
var ProductTable = &productTableType{
	s: parse.StructInfo{Type: "Product", SQLSchema: "", SQLName: "product", Fields: []parse.FieldInfo{{Name: "ProductID", Type: "int64", Column: "product_id"}, {Name: "PartyID", Type: "int64", Column: "party_id"}, {Name: "Serial", Type: "sql.NullInt64", Column: "serial"}, {Name: "Place", Type: "int", Column: "place"}, {Name: "ProductTypeName", Type: "sql.NullString", Column: "product_type_name"}, {Name: "Note", Type: "sql.NullString", Column: "note"}, {Name: "IFMinus20", Type: "sql.NullFloat64", Column: "i_f_minus20"}, {Name: "IFPlus20", Type: "sql.NullFloat64", Column: "i_f_plus20"}, {Name: "IFPlus50", Type: "sql.NullFloat64", Column: "i_f_plus50"}, {Name: "ISMinus20", Type: "sql.NullFloat64", Column: "i_s_minus20"}, {Name: "ISPlus20", Type: "sql.NullFloat64", Column: "i_s_plus20"}, {Name: "ISPlus50", Type: "sql.NullFloat64", Column: "i_s_plus50"}, {Name: "I13", Type: "sql.NullFloat64", Column: "i13"}, {Name: "I24", Type: "sql.NullFloat64", Column: "i24"}, {Name: "I35", Type: "sql.NullFloat64", Column: "i35"}, {Name: "I26", Type: "sql.NullFloat64", Column: "i26"}, {Name: "I17", Type: "sql.NullFloat64", Column: "i17"}, {Name: "NotMeasured", Type: "sql.NullFloat64", Column: "not_measured"}, {Name: "Firmware", Type: "[]uint8", Column: "firmware"}, {Name: "Production", Type: "bool", Column: "production"}, {Name: "OldProductID", Type: "sql.NullString", Column: "old_product_id"}, {Name: "OldSerial", Type: "sql.NullInt64", Column: "old_serial"}, {Name: "PointsMethod", Type: "sql.NullInt64", Column: "points_method"}}, PKFieldIndex: 0},
	z: new(Product).Values(),
}

// String returns a string representation of this struct or record.
func (s Product) String() string {
	res := make([]string, 23)
	res[0] = "ProductID: " + reform.Inspect(s.ProductID, true)
	res[1] = "PartyID: " + reform.Inspect(s.PartyID, true)
	res[2] = "Serial: " + reform.Inspect(s.Serial, true)
	res[3] = "Place: " + reform.Inspect(s.Place, true)
	res[4] = "ProductTypeName: " + reform.Inspect(s.ProductTypeName, true)
	res[5] = "Note: " + reform.Inspect(s.Note, true)
	res[6] = "IFMinus20: " + reform.Inspect(s.IFMinus20, true)
	res[7] = "IFPlus20: " + reform.Inspect(s.IFPlus20, true)
	res[8] = "IFPlus50: " + reform.Inspect(s.IFPlus50, true)
	res[9] = "ISMinus20: " + reform.Inspect(s.ISMinus20, true)
	res[10] = "ISPlus20: " + reform.Inspect(s.ISPlus20, true)
	res[11] = "ISPlus50: " + reform.Inspect(s.ISPlus50, true)
	res[12] = "I13: " + reform.Inspect(s.I13, true)
	res[13] = "I24: " + reform.Inspect(s.I24, true)
	res[14] = "I35: " + reform.Inspect(s.I35, true)
	res[15] = "I26: " + reform.Inspect(s.I26, true)
	res[16] = "I17: " + reform.Inspect(s.I17, true)
	res[17] = "NotMeasured: " + reform.Inspect(s.NotMeasured, true)
	res[18] = "Firmware: " + reform.Inspect(s.Firmware, true)
	res[19] = "Production: " + reform.Inspect(s.Production, true)
	res[20] = "OldProductID: " + reform.Inspect(s.OldProductID, true)
	res[21] = "OldSerial: " + reform.Inspect(s.OldSerial, true)
	res[22] = "PointsMethod: " + reform.Inspect(s.PointsMethod, true)
	return strings.Join(res, ", ")
}

// Values returns a slice of struct or record field values.
// Returned interface{} values are never untyped nils.
func (s *Product) Values() []interface{} {
	return []interface{}{
		s.ProductID,
		s.PartyID,
		s.Serial,
		s.Place,
		s.ProductTypeName,
		s.Note,
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
		s.Firmware,
		s.Production,
		s.OldProductID,
		s.OldSerial,
		s.PointsMethod,
	}
}

// Pointers returns a slice of pointers to struct or record fields.
// Returned interface{} values are never untyped nils.
func (s *Product) Pointers() []interface{} {
	return []interface{}{
		&s.ProductID,
		&s.PartyID,
		&s.Serial,
		&s.Place,
		&s.ProductTypeName,
		&s.Note,
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
		&s.Firmware,
		&s.Production,
		&s.OldProductID,
		&s.OldSerial,
		&s.PointsMethod,
	}
}

// View returns View object for that struct.
func (s *Product) View() reform.View {
	return ProductTable
}

// TableXY returns TableXY object for that record.
func (s *Product) Table() reform.Table {
	return ProductTable
}

// PKValue returns a value of primary key for that record.
// Returned interface{} value is never untyped nil.
func (s *Product) PKValue() interface{} {
	return s.ProductID
}

// PKPointer returns a pointer to primary key field for that record.
// Returned interface{} value is never untyped nil.
func (s *Product) PKPointer() interface{} {
	return &s.ProductID
}

// HasPK returns true if record has non-zero primary key set, false otherwise.
func (s *Product) HasPK() bool {
	return s.ProductID != ProductTable.z[ProductTable.s.PKFieldIndex]
}

// SetPK sets record primary key.
func (s *Product) SetPK(pk interface{}) {
	if i64, ok := pk.(int64); ok {
		s.ProductID = int64(i64)
	} else {
		s.ProductID = pk.(int64)
	}
}

// check interfaces
var (
	_ reform.View   = ProductTable
	_ reform.Struct = (*Product)(nil)
	_ reform.Table  = ProductTable
	_ reform.Record = (*Product)(nil)
	_ fmt.Stringer  = (*Product)(nil)
)

func init() {
	parse.AssertUpToDate(&ProductTable.s, new(Product))
}
