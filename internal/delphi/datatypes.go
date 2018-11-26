package delphi

import (
	"fmt"
	r "reflect"
	"time"

	"github.com/pkg/errors"
)

func NewDataTypes(types []r.Type, ta typesNames) *DataTypesSrc {
	src := &DataTypesSrc{
		unitName:   "server_data_types",
		implUses:   []string{"Rest.Json"},
		typesNames: ta,
	}
	for _, t := range types {
		src.addType(t)
	}
	return src
}

type DataTypesSrc struct {
	types    []dataTypeInfo
	unitName string
	interfaceUses,
	implUses []string
	typesNames typesNames
}

type dataTypeInfo struct {
	name   string
	fields []dataField
}

type dataField struct {
	name,
	typeName string
	isClass,
	isArray bool
}

func (x *DataTypesSrc) delphiClassName(t r.Type) string {
	return delphiClassName(x.typesNames, t)
}

func (x *DataTypesSrc) addType(t r.Type) {
	if t == tTime || t.Kind() != r.Struct {
		return
	}

	typeName := x.delphiClassName(t)

	for _, a := range x.types {
		if a.name == typeName {
			return
		}
	}
	ti := dataTypeInfo{
		name: typeName,
	}
	fields := x.listFields(t)
	for _, f := range fields {

		switch f.Type.Kind() {
		case r.Struct:
			x.addType(f.Type)
		case r.Slice, r.Array:
			x.addType(f.Type.Elem())
		}
		newF, err := x.newField(f)
		if err != nil {
			panic(fmt.Sprintf("type %q: %v", t.Name(), err))
		}
		ti.fields = append(ti.fields, newF)
	}
	x.types = append(x.types, ti)
	return
}

var tTime = r.TypeOf((*time.Time)(nil)).Elem()

func (x *DataTypesSrc) newField(structField r.StructField) (dataField, error) {
	f := dataField{
		name: structField.Name,
	}

	kind := structField.Type.Kind()

	if structField.Type == r.TypeOf((*time.Time)(nil)).Elem() {
		f.typeName = "TDateTime"
		return f, nil
	}

	switch kind {

	case r.Float32:
		f.typeName = "Single"

	case r.Float64:
		f.typeName = "Double"

	case r.Int:
		f.typeName = "Integer"

	case r.Uint8:
		f.typeName = "Byte"

	case r.Uint16:
		f.typeName = "Word"

	case r.Uint32:
		f.typeName = "Cardinal"

	case r.Uint64:
		f.typeName = "UInt64"

	case r.Int8:
		f.typeName = "ShortInt"

	case r.Int16:
		f.typeName = "SmallInt"

	case r.Int32:
		f.typeName = "Integer"

	case r.Int64:
		f.typeName = "Int64"

	case r.Bool:
		f.typeName = "Boolean"

	case r.String:
		f.typeName = "string"

	case r.Slice, r.Array:
		f.isArray = true
		f.typeName = structField.Type.Elem().Name()
		if structField.Type.Elem().Kind() == r.Struct {
			f.typeName = x.delphiClassName(structField.Type.Elem())
		}
	case r.Struct:
		f.isClass = true
		f.typeName = x.delphiClassName(structField.Type)

	default:
		return f, errors.Errorf("type not supported: %q, dataField %q", structField.Type.Name(), structField.Name)
	}

	return f, nil
}

func (x *DataTypesSrc) listFields(t r.Type) (fields []r.StructField) {
	num := t.NumField()
	for i := 0; i < num; i++ {
		f := t.Field(i)

		if f.Anonymous {
			fields = append(fields, x.listFields(f.Type)...)
		} else {
			fields = append(fields, f)
		}
	}
	return
}

func (x *dataTypeInfo) hasClassField() bool {
	for _, a := range x.fields {
		if a.isClass {
			return true
		}
	}
	return false
}
