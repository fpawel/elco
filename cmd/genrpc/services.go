package main

import (
	"fmt"
	"log"
	r "reflect"
)

type ServicesUnit struct {
	TypesUnit *TypesUnit
	Name      string
	services  []service
}

type service struct {
	serviceName string
	methods     []delphiServiceMethod
}

func NewServicesUnit(name string, types []r.Type) *ServicesUnit {
	src := &ServicesUnit{
		Name:      name,
		TypesUnit: new(TypesUnit),
	}
	for _, t := range types {
		src.add(t)
	}
	return src
}

func (x *ServicesUnit) add(serviceType r.Type) {
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

func (x *ServicesUnit) method(met r.Method) (m delphiServiceMethod) {
	m.methodName = met.Name
	argType := met.Type.In(1)

	fmt.Println("method:", met.Name)

	switch argType.Kind() {
	case r.Array:
		for i := 0; i < argType.Len(); i++ {
			m.addParam(fmt.Sprintf("param%d", i+1), argType.Elem(), x.TypesUnit)
		}
	case r.Struct:
		for _, f := range listTypeFields(argType) {
			m.namedParams = true
			m.addParam(f.Name, f.Type, x.TypesUnit)
		}

	default:
		panic(fmt.Sprintf("%v: %v: must be array or struct", met, argType))
	}

	returnType := met.Type.In(2).Elem()

	switch returnType.Kind() {
	case r.Slice:
		t, err := x.TypesUnit.add(returnType)
		if err != nil {
			log.Fatalln("method return type:", met.Name, "error:", err)
		}
		m.ret = &t
	case r.Struct:
		if returnType.NumField() > 0 {
			t, err := x.TypesUnit.add(returnType)
			if err != nil {
				log.Fatalln("method return type:", met.Name, "error:", err)
			}
			m.ret = &t
		}
	default:
		t, err := x.TypesUnit.add(returnType)
		if err != nil {
			log.Fatalln("method return type:", met.Name, "error:", err)
		}
		m.ret = &t
	}
	return
}
