package delphi

import (
	"fmt"
	"io"
	r "reflect"
)

type ServicesSrc struct {
	unitName string
	interfaceUses,
	implementationUses []string
	dataTypes *DataTypesSrc
}

type service struct {
	serviceName string
}

type method struct {
	methodName  string
	namedParams bool
	params      []param
	retType     r.Type
	retArray    bool
}

type param struct {
	name, typeName string
}

func ServicesUnit(types []r.Type, ta typesNames, w io.Writer) {
	src := ServicesSrc{
		unitName:           "services",
		implementationUses: []string{"Rest.Json"},
		dataTypes: &DataTypesSrc{
			unitName:   "server_data_types",
			implUses:   []string{"Rest.Json"},
			typesNames: ta,
		},
	}
	for _, t := range types {
		src.addType(t)
	}
	WriteDataTypesUnit(w, src.dataTypes)
}

func (x *ServicesSrc) addType(serviceType r.Type) {

	fmt.Println(serviceType.Elem().Name())
	for nMethod := 0; nMethod < serviceType.NumMethod(); nMethod++ {
		met := serviceType.Method(nMethod)
		fmt.Println(serviceType.Name() + met.Name)
		//x.funcAnalyse(serviceType.Elem().Name() + met.Name, met.Type)
	}
	return
}

func (x *ServicesSrc) method(methodType r.Type) (m method) {

	argType := methodType.In(1)
	//fmt.Println("\t", argType.String() )

	if argType.Kind() == r.Array {
		for i := 0; i < argType.Len(); i++ {
			m.params = append(m.params, param{
				name:     fmt.Sprintf("param%d", i+1),
				typeName: argType.Elem().Name(),
			})
		}
	} else {
		if argType.Kind() != r.Struct {
			panic(fmt.Sprintf("%v: must be array or struct", argType.Kind()))
		}
		m.namedParams = true
		for i := 0; i < argType.NumField(); i++ {
			f := argType.Field(i)
			typeName := delphiPlainOldTypeName(f)
			if typeName == "" {
				panic(fmt.Sprintf("%v: must be POD", f))
			}
			m.params = append(m.params, param{
				name:     f.Name,
				typeName: delphiPlainOldTypeName(f),
			})
		}
	}

	returnType := methodType.In(2).Elem()

	if returnType.Kind() == r.Slice {
		returnType = returnType.Elem()
	}

	if returnType.Kind() == r.Struct && returnType.NumField() > 0 {
		x.dataTypes.addType(returnType)
	}
}
