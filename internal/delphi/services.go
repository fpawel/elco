package delphi

import (
	"fmt"
	"io"
	r "reflect"
)

type ServicesSrc struct {
	unitName string
	interfaceUses,
	implUses []string
	dataTypes *TypesSrc
	services  []service
	pipe      string
}

type service struct {
	serviceName string
	methods     []method
}

type method struct {
	methodName    string
	namedParams   bool
	params        []param
	retDelphiType string
	retArray      bool
	retPODType    bool
	procedure     bool
}

type param struct {
	name, typeName string
}

func ServicesUnit(pipe string, types []r.Type, ta typesNames, wServices, wTypes io.Writer) {
	src := ServicesSrc{
		pipe:          pipe,
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
		src.addService(t)
	}
	src.dataTypes.WriteUnit(wTypes)
	src.WriteUnit(wServices)
}
func (x *ServicesSrc) pipeStr() string {
	return "'" + x.pipe + "'"
}

func (x *ServicesSrc) addService(serviceType r.Type) {
	srv := service{
		serviceName: serviceType.Elem().Name(),
	}
	for nMethod := 0; nMethod < serviceType.NumMethod(); nMethod++ {
		met := serviceType.Method(nMethod)
		srv.methods = append(srv.methods, x.method(met))
	}
	x.services = append(x.services, srv)
	return
}

func (x *ServicesSrc) method(met r.Method) (m method) {
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
			panic(fmt.Sprintf("%v: %v: must be array or struct", met, argType))
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
		returnType = returnType.Elem()
		x.dataTypes.addType(returnType)
		m.retArray = true
		m.retPODType = len(delphiPlainOldTypeName(returnType)) > 0
		m.retDelphiType = delphiTypeName(x.dataTypes.typesNames, returnType)
	case r.Struct:
		if returnType.NumField() == 0 {
			m.procedure = true
		} else {
			x.dataTypes.addType(returnType)
			m.retDelphiType = delphiTypeName(x.dataTypes.typesNames, returnType)
			m.retPODType = len(delphiPlainOldTypeName(returnType)) > 0
		}
	default:
		m.retDelphiType = delphiPlainOldTypeName(returnType)
		m.retPODType = true
	}
	//m.writebody(os.Stdout)
	return
}

func genSetField(paramName string) string {
	return fmt.Sprintf("SuperObject_SetField(req, '%s', %s)",
		paramName, paramName)
}

func (x method) remoteMethod(serviceName string) string {
	return fmt.Sprintf("'%s.%s'", serviceName, x.methodName)
}
