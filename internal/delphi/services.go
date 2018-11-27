package delphi

import (
	"fmt"
	"io"
	r "reflect"
	"strings"
)

type ServicesSrc struct {
	unitName string
	interfaceUses,
	implUses []string
	dataTypes *TypesSrc
	methods   []method
}

type method struct {
	serviceName   string
	methodName    string
	namedParams   bool
	params        []param
	retType       r.Type
	retDelphiType string
	retArray      bool
	procedure     bool
}

type param struct {
	name, typeName string
}

func ServicesUnit(types []r.Type, ta typesNames, wServices, wTypes io.Writer) {
	src := ServicesSrc{
		unitName:      "services",
		interfaceUses: []string{"server_data_types", "pipe", "superobject"},
		implUses:      []string{"Rest.Json"},
		dataTypes: &TypesSrc{
			unitName:   "server_data_types",
			implUses:   []string{"Rest.Json"},
			typesNames: ta,
		},
	}
	for _, t := range types {
		src.addType(t)
	}
	src.dataTypes.WriteUnit(wTypes)
	src.WriteUnit(wServices)
}

func (x *ServicesSrc) addType(serviceType r.Type) {

	fmt.Println(serviceType.Elem().Name())
	for nMethod := 0; nMethod < serviceType.NumMethod(); nMethod++ {
		met := serviceType.Method(nMethod)
		//fmt.Println(serviceType.Name() + met.Name)
		//x.funcAnalyse(serviceType.Elem().Name() + met.Name, met.Type)

		x.methods = append(x.methods, x.method(serviceType.Elem().Name(), met))

	}
	return
}

func (x *ServicesSrc) method(serviceName string, met r.Method) (m method) {
	m.serviceName = serviceName
	m.methodName = met.Name

	argType := met.Type.In(1)
	//fmt.Println("\t", argType.String() )

	if argType.Kind() == r.Array {
		for i := 0; i < argType.Len(); i++ {
			m.params = append(m.params, param{
				name:     fmt.Sprintf("param%d", i+1),
				typeName: delphiPlainOldTypeName(argType.Elem()),
			})
		}
	} else {
		if argType.Kind() != r.Struct {
			panic(fmt.Sprintf("%v: must be array or struct", argType.Kind()))
		}
		m.namedParams = true
		for i := 0; i < argType.NumField(); i++ {
			f := argType.Field(i)
			typeName := delphiPlainOldTypeName(f.Type)
			if typeName == "" {
				panic(fmt.Sprintf("%v: must be POD", f))
			}
			m.params = append(m.params, param{
				name:     f.Name,
				typeName: delphiPlainOldTypeName(f.Type),
			})
		}
	}

	returnType := met.Type.In(2).Elem()

	switch returnType.Kind() {
	case r.Slice:
		m.retType = returnType.Elem()
		m.retArray = true
		m.retDelphiType = "TArray<" + delphiTypeName(x.dataTypes.typesNames, m.retType) + ">"
	case r.Struct:
		if returnType.NumField() == 0 {
			m.procedure = true
		} else {
			x.dataTypes.addType(returnType)
			m.retType = returnType
			m.retDelphiType = delphiTypeName(x.dataTypes.typesNames, m.retType)
		}
	default:
		m.retType = returnType
		m.retDelphiType = delphiPlainOldTypeName(m.retType)
	}
	//m.writebody(os.Stdout)
	return
}

func (x method) signatureParams() string {
	var s []string
	for _, p := range x.params {
		s = append(s, fmt.Sprintf("%s: %s", p.name, p.typeName))
	}
	return strings.Join(s, "; ")
}

func genSetField(paramName string) string {
	return fmt.Sprintf("SuperObject_SetField(req, '%s', %s)",
		paramName, paramName)
}

func (x method) genMethod() string {
	return fmt.Sprintf("'%s.%s'", x.serviceName, x.methodName)
}

//func (x method)  signature() (s string){
//	if x.procedure {
//		s += "procedure "
//	} else {
//		s += "function "
//	}
//	s += x.serviceName + "_" + x.methodName + "(pipe_conn:TPipe; " + x.signatureParams() + ")"
//	if !x.procedure {
//		s += ": "+ x.retDelphiType
//	}
//	return
//}
