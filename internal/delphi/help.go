package delphi

import (
	"github.com/pkg/errors"
	r "reflect"
	"time"
)

type typesNames = map[string]string

func delphiTypeName(m typesNames, structField r.StructField) (dataField, error) {
	f := dataField{
		name: structField.Name,
	}

	kind := structField.Type.Kind()

	if podName := delphiPlainOldTypeName(structField); podName != "" {
		f.typeName = podName
		return f, nil
	}

	switch kind {

	case r.Slice, r.Array:
		f.isArray = true
		f.typeName = structField.Type.Elem().Name()
		if structField.Type.Elem().Kind() == r.Struct {
			f.typeName = delphiClassName(m, structField.Type.Elem())
		}
	case r.Struct:
		f.isClass = true
		f.typeName = delphiClassName(m, structField.Type)

	default:
		return f, errors.Errorf("type not supported: %q, dataField %q", structField.Type.Name(), structField.Name)
	}

	return f, nil
}

func delphiPlainOldTypeName(structField r.StructField) string {
	kind := structField.Type.Kind()

	if structField.Type == r.TypeOf((*time.Time)(nil)).Elem() {
		return "TDateTime"
	}

	switch kind {

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

func delphiClassName(m typesNames, t r.Type) string {
	typeName := t.Name()
	if typeName == "" {
		return "T" + m[t.String()]
	}
	if typeAlias, hasTypeAlias := m[typeName]; hasTypeAlias {
		return "T" + typeAlias
	}
	return "T" + typeName
}
