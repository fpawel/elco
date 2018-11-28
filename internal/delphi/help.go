package delphi

import (
	"fmt"
	"github.com/pkg/errors"
	r "reflect"
	"strings"
	"time"
)

type typesNames = map[string]string

func uses(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return "uses " + strings.Join(s, ", ") + ";"
}

func newField(fieldName string, fieldType r.Type, m typesNames) (dataField, error) {
	f := dataField{
		name: fieldName,
	}

	if podName := delphiPlainOldTypeName(fieldType); podName != "" {
		f.typeName = podName
		return f, nil
	}

	switch fieldType.Kind() {

	case r.Slice, r.Array:
		f.isArray = true
		f.typeName = delphiTypeName(m, fieldType.Elem())
	case r.Struct:
		f.isClass = true
		f.typeName = delphiTypeName(m, fieldType)
	default:
		return f, errors.Errorf("type not supported: %q, dataField %q", fieldType.Name(), fieldName)
	}

	return f, nil
}

func delphiTypeName(m typesNames, t r.Type) string {
	s := delphiPlainOldTypeName(t)
	if len(s) > 0 {
		return s
	}
	return structNameToDelphiClassName(m, t)
}

func delphiPlainOldTypeName(t r.Type) string {

	if t == r.TypeOf((*time.Time)(nil)).Elem() {
		return "TDateTime"
	}

	switch t.Kind() {

	case r.Float32:
		return "Single"

	case r.Float64:
		return "Double"

	case r.Int:
		return "Integer"

	case r.Uint8:
		return "Byte"

	case r.Uint16:
		return "Word"

	case r.Uint32:
		return "Cardinal"

	case r.Uint64:
		return "UInt64"

	case r.Int8:
		return "ShortInt"

	case r.Int16:
		return "SmallInt"

	case r.Int32:
		return "Integer"

	case r.Int64:
		return "Int64"

	case r.Bool:
		return "Boolean"

	case r.String:
		return "string"

	default:
		return ""
	}
}

func structNameToDelphiClassName(m typesNames, t r.Type) string {
	typeName := t.Name()
	if typeName == "" {
		panic(fmt.Sprintf("rmpty type name %+v", t))
	}
	if typeAlias, hasTypeAlias := m[typeName]; hasTypeAlias {
		return "T" + typeAlias
	}
	return "T" + typeName
}
