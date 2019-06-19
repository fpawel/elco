package delphirpc

import (
	"fmt"
	"log"
	r "reflect"
	"time"
)

//func NewTypes(types []r.Type, ta typesNames) *TypesSrc {
//	src := &TypesSrc{
//		unitName:   "server_data_types",
//		implUses:   []string{"Rest.Json"},
//		typesNames: ta,
//	}
//	for _, t := range types {
//		src.addType(t)
//	}
//	return src
//}

type TypesSrc struct {
	types    []typeInfo
	unitName string
	interfaceUses,
	implUses []string
	typesNames typesNames
}

type typeInfo struct {
	name   string
	fields []dataField
}

type dataField struct {
	name,
	typeName string
	isArray bool
}

func (x *TypesSrc) addType(t r.Type) {
	if t == r.TypeOf((*time.Time)(nil)).Elem() || t.Kind() != r.Struct {
		return
	}

	typeName := delphiTypeName(x.typesNames, t)

	for _, a := range x.types {
		if a.name == typeName {
			return
		}
	}

	fmt.Println(t.Name(), t, typeName)

	ti := typeInfo{
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
		newF, err := newField(f.Name, f.Type, x.typesNames)
		if err != nil {
			log.Panicf("type %q: %v\n", t.Name(), err)
		}
		ti.fields = append(ti.fields, newF)
	}
	x.types = append(x.types, ti)
	return
}

func (x *TypesSrc) listFields(t r.Type) (fields []r.StructField) {
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

func (x dataField) declType() string {
	if x.isArray {
		return fmt.Sprintf("TArray<%s>", x.typeName)
	}
	return x.typeName
}
