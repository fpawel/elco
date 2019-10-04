package main

import (
	"fmt"
	r "reflect"
	"time"
)

type TypesUnit []delphiType

func (x *TypesUnit) add(t r.Type) (delphiType, error) {

	if t == r.TypeOf((*time.Time)(nil)).Elem() {
		return delphiType{name: "TDateTime", kind: delphiPOD}, nil
	}

	newType, err := newDelphiType(t)

	if err != nil {
		return delphiType{}, err
	}

	for _, a := range *x {
		if a.TypeName() == newType.TypeName() {
			return newType, nil
		}
	}

	switch t.Kind() {
	case r.Struct:
		for _, f := range listTypeFields(t) {
			_, err := x.add(f.Type)
			if err != nil {
				return delphiType{}, err
			}
		}
	case r.Array, r.Slice:
		_, err := x.add(t.Elem())
		if err != nil {
			return delphiType{}, err
		}
	}

	*x = append(*x, newType)

	fmt.Println("\t\t    Go type:", t)
	fmt.Println("\t\tdelphi type:", newType.TypeName())
	return newType, nil
}

func listTypeFields(t r.Type) (fields []r.StructField) {
	num := t.NumField()
	for i := 0; i < num; i++ {
		f := t.Field(i)

		if f.Anonymous {
			fields = append(fields, listTypeFields(f.Type)...)
		} else {
			fields = append(fields, f)
		}
	}
	return
}

type delphiType struct {
	name   string
	fields map[string]delphiType
	elem   *delphiType
	kind   delphiTypeKind
}

type delphiTypeKind int

const (
	delphiPOD delphiTypeKind = iota
	delphiRecord
	delphiArray
)

func (x delphiType) TypeName() string {
	switch x.kind {
	case delphiArray:
		return fmt.Sprintf("TArray<%s>", x.elem.TypeName())
	case delphiRecord:
		return "T" + x.name
	default:
		return x.name
	}
}

func newDelphiType(t r.Type) (delphiType, error) {

	pod := func(name string) (delphiType, error) {
		return delphiType{name: name, kind: delphiPOD}, nil
	}

	if t == r.TypeOf((*time.Time)(nil)).Elem() {
		return pod("TDateTime")
	}

	switch t.Kind() {

	case r.Float32:
		return pod("Single")

	case r.Float64:
		return pod("Double")

	case r.Int:
		return pod("Integer")

	case r.Uint8:
		return pod("Byte")

	case r.Uint16:
		return pod("Word")

	case r.Uint32:
		return pod("Cardinal")

	case r.Uint64:
		return pod("UInt64")

	case r.Int8:
		return pod("ShortInt")

	case r.Int16:
		return pod("SmallInt")

	case r.Int32:
		return pod("Integer")

	case r.Int64:
		return pod("Int64")

	case r.Bool:
		return pod("Boolean")

	case r.String:
		return pod("string")

	case r.Array, r.Slice:
		elem, err := newDelphiType(t.Elem())
		if err != nil {
			return delphiType{}, err
		}
		return delphiType{
			name: t.Name() + "Array",
			kind: delphiArray,
			elem: &elem,
		}, nil

	case r.Struct:
		x := delphiType{
			name:   t.Name(),
			kind:   delphiRecord,
			fields: make(map[string]delphiType),
		}

		for _, f := range listTypeFields(t) {
			t, err := newDelphiType(f.Type)
			if err != nil {
				return delphiType{}, err
			}
			x.fields[f.Name] = t
		}

		return x, nil

	default:
		return delphiType{}, fmt.Errorf("bad type: %v", t)
	}
}
