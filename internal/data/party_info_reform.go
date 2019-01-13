// Code generated by gopkg.in/reform.v1. DO NOT EDIT.

package data

import (
	"fmt"
	"strings"

	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/parse"
)

type partyInfoTableType struct {
	s parse.StructInfo
	z []interface{}
}

// Schema returns a schema name in SQL database ("").
func (v *partyInfoTableType) Schema() string {
	return v.s.SQLSchema
}

// Name returns a view or table name in SQL database ("party_info").
func (v *partyInfoTableType) Name() string {
	return v.s.SQLName
}

// Columns returns a new slice of column names for that view or table in SQL database.
func (v *partyInfoTableType) Columns() []string {
	return []string{"party_id", "created_at", "updated_at", "product_type_name", "concentration1", "concentration2", "concentration3", "note", "last", "min_fon", "max_fon", "max_d_fon", "min_k_sens20", "max_k_sens20", "min_k_sens50", "max_k_sens50", "min_d_temp", "max_d_temp", "max_d_not_measured"}
}

// NewStruct makes a new struct for that view or table.
func (v *partyInfoTableType) NewStruct() reform.Struct {
	return new(PartyInfo)
}

// NewRecord makes a new record for that table.
func (v *partyInfoTableType) NewRecord() reform.Record {
	return new(PartyInfo)
}

// PKColumnIndex returns an index of primary key column for that table in SQL database.
func (v *partyInfoTableType) PKColumnIndex() uint {
	return uint(v.s.PKFieldIndex)
}

// PartyInfoTable represents party_info view or table in SQL database.
var PartyInfoTable = &partyInfoTableType{
	s: parse.StructInfo{Type: "PartyInfo", SQLSchema: "", SQLName: "party_info", Fields: []parse.FieldInfo{{Name: "PartyID", Type: "int64", Column: "party_id"}, {Name: "CreatedAt", Type: "time.Time", Column: "created_at"}, {Name: "UpdatedAt", Type: "time.Time", Column: "updated_at"}, {Name: "ProductTypeName", Type: "string", Column: "product_type_name"}, {Name: "Concentration1", Type: "float64", Column: "concentration1"}, {Name: "Concentration2", Type: "float64", Column: "concentration2"}, {Name: "Concentration3", Type: "float64", Column: "concentration3"}, {Name: "Note", Type: "sql.NullString", Column: "note"}, {Name: "Last", Type: "bool", Column: "last"}, {Name: "MinFon", Type: "sql.NullFloat64", Column: "min_fon"}, {Name: "MaxFon", Type: "sql.NullFloat64", Column: "max_fon"}, {Name: "MaxDFon", Type: "sql.NullFloat64", Column: "max_d_fon"}, {Name: "MinKSens20", Type: "sql.NullFloat64", Column: "min_k_sens20"}, {Name: "MaxKSens20", Type: "sql.NullFloat64", Column: "max_k_sens20"}, {Name: "MinKSens50", Type: "sql.NullFloat64", Column: "min_k_sens50"}, {Name: "MaxKSens50", Type: "sql.NullFloat64", Column: "max_k_sens50"}, {Name: "MinDTemp", Type: "sql.NullFloat64", Column: "min_d_temp"}, {Name: "MaxDTemp", Type: "sql.NullFloat64", Column: "max_d_temp"}, {Name: "MaxDNotMeasured", Type: "sql.NullFloat64", Column: "max_d_not_measured"}}, PKFieldIndex: 0},
	z: new(PartyInfo).Values(),
}

// String returns a string representation of this struct or record.
func (s PartyInfo) String() string {
	res := make([]string, 19)
	res[0] = "PartyID: " + reform.Inspect(s.PartyID, true)
	res[1] = "CreatedAt: " + reform.Inspect(s.CreatedAt, true)
	res[2] = "UpdatedAt: " + reform.Inspect(s.UpdatedAt, true)
	res[3] = "ProductTypeName: " + reform.Inspect(s.ProductTypeName, true)
	res[4] = "Concentration1: " + reform.Inspect(s.Concentration1, true)
	res[5] = "Concentration2: " + reform.Inspect(s.Concentration2, true)
	res[6] = "Concentration3: " + reform.Inspect(s.Concentration3, true)
	res[7] = "Note: " + reform.Inspect(s.Note, true)
	res[8] = "Last: " + reform.Inspect(s.Last, true)
	res[9] = "MinFon: " + reform.Inspect(s.MinFon, true)
	res[10] = "MaxFon: " + reform.Inspect(s.MaxFon, true)
	res[11] = "MaxDFon: " + reform.Inspect(s.MaxDFon, true)
	res[12] = "MinKSens20: " + reform.Inspect(s.MinKSens20, true)
	res[13] = "MaxKSens20: " + reform.Inspect(s.MaxKSens20, true)
	res[14] = "MinKSens50: " + reform.Inspect(s.MinKSens50, true)
	res[15] = "MaxKSens50: " + reform.Inspect(s.MaxKSens50, true)
	res[16] = "MinDTemp: " + reform.Inspect(s.MinDTemp, true)
	res[17] = "MaxDTemp: " + reform.Inspect(s.MaxDTemp, true)
	res[18] = "MaxDNotMeasured: " + reform.Inspect(s.MaxDNotMeasured, true)
	return strings.Join(res, ", ")
}

// Values returns a slice of struct or record field values.
// Returned interface{} values are never untyped nils.
func (s *PartyInfo) Values() []interface{} {
	return []interface{}{
		s.PartyID,
		s.CreatedAt,
		s.UpdatedAt,
		s.ProductTypeName,
		s.Concentration1,
		s.Concentration2,
		s.Concentration3,
		s.Note,
		s.Last,
		s.MinFon,
		s.MaxFon,
		s.MaxDFon,
		s.MinKSens20,
		s.MaxKSens20,
		s.MinKSens50,
		s.MaxKSens50,
		s.MinDTemp,
		s.MaxDTemp,
		s.MaxDNotMeasured,
	}
}

// Pointers returns a slice of pointers to struct or record fields.
// Returned interface{} values are never untyped nils.
func (s *PartyInfo) Pointers() []interface{} {
	return []interface{}{
		&s.PartyID,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.ProductTypeName,
		&s.Concentration1,
		&s.Concentration2,
		&s.Concentration3,
		&s.Note,
		&s.Last,
		&s.MinFon,
		&s.MaxFon,
		&s.MaxDFon,
		&s.MinKSens20,
		&s.MaxKSens20,
		&s.MinKSens50,
		&s.MaxKSens50,
		&s.MinDTemp,
		&s.MaxDTemp,
		&s.MaxDNotMeasured,
	}
}

// View returns View object for that struct.
func (s *PartyInfo) View() reform.View {
	return PartyInfoTable
}

// TableXY returns TableXY object for that record.
func (s *PartyInfo) Table() reform.Table {
	return PartyInfoTable
}

// PKValue returns a value of primary key for that record.
// Returned interface{} value is never untyped nil.
func (s *PartyInfo) PKValue() interface{} {
	return s.PartyID
}

// PKPointer returns a pointer to primary key field for that record.
// Returned interface{} value is never untyped nil.
func (s *PartyInfo) PKPointer() interface{} {
	return &s.PartyID
}

// HasPK returns true if record has non-zero primary key set, false otherwise.
func (s *PartyInfo) HasPK() bool {
	return s.PartyID != PartyInfoTable.z[PartyInfoTable.s.PKFieldIndex]
}

// SetPK sets record primary key.
func (s *PartyInfo) SetPK(pk interface{}) {
	if i64, ok := pk.(int64); ok {
		s.PartyID = int64(i64)
	} else {
		s.PartyID = pk.(int64)
	}
}

// check interfaces
var (
	_ reform.View   = PartyInfoTable
	_ reform.Struct = (*PartyInfo)(nil)
	_ reform.Table  = PartyInfoTable
	_ reform.Record = (*PartyInfo)(nil)
	_ fmt.Stringer  = (*PartyInfo)(nil)
)

func init() {
	parse.AssertUpToDate(&PartyInfoTable.s, new(PartyInfo))
}
