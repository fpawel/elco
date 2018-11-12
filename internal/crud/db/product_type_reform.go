// Code generated by gopkg.in/reform.v1. DO NOT EDIT.

package db

import (
	"fmt"
	"strings"

	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/parse"
)

type productTypeTableType struct {
	s parse.StructInfo
	z []interface{}
}

// Schema returns a schema name in SQL database ("").
func (v *productTypeTableType) Schema() string {
	return v.s.SQLSchema
}

// Name returns a view or table name in SQL database ("product_type").
func (v *productTypeTableType) Name() string {
	return v.s.SQLName
}

// Columns returns a new slice of column names for that view or table in SQL database.
func (v *productTypeTableType) Columns() []string {
	return []string{"product_type_name", "display_name", "gas_name", "units_name", "scale", "noble_metal_content", "lifetime_months", "lc64", "points_method", "max_fon", "max_d_fon", "min_k_sens20", "max_k_sens20", "min_d_temp", "max_d_temp", "min_k_sens50", "max_k_sens50", "max_d_not_measured"}
}

// NewStruct makes a new struct for that view or table.
func (v *productTypeTableType) NewStruct() reform.Struct {
	return new(ProductType)
}

// NewRecord makes a new record for that table.
func (v *productTypeTableType) NewRecord() reform.Record {
	return new(ProductType)
}

// PKColumnIndex returns an index of primary key column for that table in SQL database.
func (v *productTypeTableType) PKColumnIndex() uint {
	return uint(v.s.PKFieldIndex)
}

// ProductTypeTable represents product_type view or table in SQL database.
var ProductTypeTable = &productTypeTableType{
	s: parse.StructInfo{Type: "ProductType", SQLSchema: "", SQLName: "product_type", Fields: []parse.FieldInfo{{Name: "ProductTypeName", Type: "string", Column: "product_type_name"}, {Name: "DisplayName", Type: "*string", Column: "display_name"}, {Name: "GasName", Type: "string", Column: "gas_name"}, {Name: "UnitsName", Type: "string", Column: "units_name"}, {Name: "Scale", Type: "float64", Column: "scale"}, {Name: "NobleMetalContent", Type: "float64", Column: "noble_metal_content"}, {Name: "LifetimeMonths", Type: "int64", Column: "lifetime_months"}, {Name: "Lc64", Type: "bool", Column: "lc64"}, {Name: "PointsMethod", Type: "int64", Column: "points_method"}, {Name: "MaxFon", Type: "*float64", Column: "max_fon"}, {Name: "MaxDFon", Type: "*float64", Column: "max_d_fon"}, {Name: "MinKSens20", Type: "*float64", Column: "min_k_sens20"}, {Name: "MaxKSens20", Type: "*float64", Column: "max_k_sens20"}, {Name: "MinDTemp", Type: "*float64", Column: "min_d_temp"}, {Name: "MaxDTemp", Type: "*float64", Column: "max_d_temp"}, {Name: "MinKSens50", Type: "*float64", Column: "min_k_sens50"}, {Name: "MaxKSens50", Type: "*float64", Column: "max_k_sens50"}, {Name: "MaxDNotMeasured", Type: "*float64", Column: "max_d_not_measured"}}, PKFieldIndex: 0},
	z: new(ProductType).Values(),
}

// String returns a string representation of this struct or record.
func (s ProductType) String() string {
	res := make([]string, 18)
	res[0] = "ProductTypeName: " + reform.Inspect(s.ProductTypeName, true)
	res[1] = "DisplayName: " + reform.Inspect(s.DisplayName, true)
	res[2] = "GasName: " + reform.Inspect(s.GasName, true)
	res[3] = "UnitsName: " + reform.Inspect(s.UnitsName, true)
	res[4] = "Scale: " + reform.Inspect(s.Scale, true)
	res[5] = "NobleMetalContent: " + reform.Inspect(s.NobleMetalContent, true)
	res[6] = "LifetimeMonths: " + reform.Inspect(s.LifetimeMonths, true)
	res[7] = "Lc64: " + reform.Inspect(s.Lc64, true)
	res[8] = "PointsMethod: " + reform.Inspect(s.PointsMethod, true)
	res[9] = "MaxFon: " + reform.Inspect(s.MaxFon, true)
	res[10] = "MaxDFon: " + reform.Inspect(s.MaxDFon, true)
	res[11] = "MinKSens20: " + reform.Inspect(s.MinKSens20, true)
	res[12] = "MaxKSens20: " + reform.Inspect(s.MaxKSens20, true)
	res[13] = "MinDTemp: " + reform.Inspect(s.MinDTemp, true)
	res[14] = "MaxDTemp: " + reform.Inspect(s.MaxDTemp, true)
	res[15] = "MinKSens50: " + reform.Inspect(s.MinKSens50, true)
	res[16] = "MaxKSens50: " + reform.Inspect(s.MaxKSens50, true)
	res[17] = "MaxDNotMeasured: " + reform.Inspect(s.MaxDNotMeasured, true)
	return strings.Join(res, ", ")
}

// Values returns a slice of struct or record field values.
// Returned interface{} values are never untyped nils.
func (s *ProductType) Values() []interface{} {
	return []interface{}{
		s.ProductTypeName,
		s.DisplayName,
		s.GasName,
		s.UnitsName,
		s.Scale,
		s.NobleMetalContent,
		s.LifetimeMonths,
		s.Lc64,
		s.PointsMethod,
		s.MaxFon,
		s.MaxDFon,
		s.MinKSens20,
		s.MaxKSens20,
		s.MinDTemp,
		s.MaxDTemp,
		s.MinKSens50,
		s.MaxKSens50,
		s.MaxDNotMeasured,
	}
}

// Pointers returns a slice of pointers to struct or record fields.
// Returned interface{} values are never untyped nils.
func (s *ProductType) Pointers() []interface{} {
	return []interface{}{
		&s.ProductTypeName,
		&s.DisplayName,
		&s.GasName,
		&s.UnitsName,
		&s.Scale,
		&s.NobleMetalContent,
		&s.LifetimeMonths,
		&s.Lc64,
		&s.PointsMethod,
		&s.MaxFon,
		&s.MaxDFon,
		&s.MinKSens20,
		&s.MaxKSens20,
		&s.MinDTemp,
		&s.MaxDTemp,
		&s.MinKSens50,
		&s.MaxKSens50,
		&s.MaxDNotMeasured,
	}
}

// View returns View object for that struct.
func (s *ProductType) View() reform.View {
	return ProductTypeTable
}

// Table returns Table object for that record.
func (s *ProductType) Table() reform.Table {
	return ProductTypeTable
}

// PKValue returns a value of primary key for that record.
// Returned interface{} value is never untyped nil.
func (s *ProductType) PKValue() interface{} {
	return s.ProductTypeName
}

// PKPointer returns a pointer to primary key field for that record.
// Returned interface{} value is never untyped nil.
func (s *ProductType) PKPointer() interface{} {
	return &s.ProductTypeName
}

// HasPK returns true if record has non-zero primary key set, false otherwise.
func (s *ProductType) HasPK() bool {
	return s.ProductTypeName != ProductTypeTable.z[ProductTypeTable.s.PKFieldIndex]
}

// SetPK sets record primary key.
func (s *ProductType) SetPK(pk interface{}) {
	if i64, ok := pk.(int64); ok {
		s.ProductTypeName = string(i64)
	} else {
		s.ProductTypeName = pk.(string)
	}
}

// check interfaces
var (
	_ reform.View   = ProductTypeTable
	_ reform.Struct = (*ProductType)(nil)
	_ reform.Table  = ProductTypeTable
	_ reform.Record = (*ProductType)(nil)
	_ fmt.Stringer  = (*ProductType)(nil)
)

func init() {
	parse.AssertUpToDate(&ProductTypeTable.s, new(ProductType))
}
