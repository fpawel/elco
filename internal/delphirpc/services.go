package delphirpc

import (
	"fmt"
	r "reflect"
)

type ServicesSrc struct {
	unitName string
	interfaceUses,
	implUses []string
	DataTypes *TypesSrc
	services  []service
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
	isArray        bool
	isStructure    bool
}

func (x param) String() string {
	if x.isArray {
		return "TArray<" + x.typeName + ">"
	}
	return x.typeName
}

func (x method) hasStructureParam() bool {
	for _, p := range x.params {
		if p.isStructure {
			return true
		}
	}
	return false
}

func (x param) setFieldInstruction() string {
	if x.isStructure {
		return fmt.Sprintf("TgoBsonSerializer.serialize(%s, s); req['%s'] := SO(s)", x.name, x.name)
	}

	return fmt.Sprintf("SuperObject_SetField(req, '%s', %s)", x.name, x.name)
}

func NewServicesSrc(unitName, dataUnitName string, types []r.Type, ta typesNames) *ServicesSrc {
	src := &ServicesSrc{
		unitName:      unitName,
		interfaceUses: []string{dataUnitName, "superobject"},
		implUses:      []string{"HttpRpcClient", "SuperObjectHelp", "Grijjy.Bson.Serialization"},
		DataTypes: &TypesSrc{
			unitName:      dataUnitName,
			interfaceUses: []string{"Grijjy.Bson", "Grijjy.Bson.Serialization"},
			typesNames:    ta,
		},
	}
	for _, t := range types {
		src.addService(t)
	}
	return src
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

			p := param{name: f.Name}

			switch f.Type.Kind() {
			case r.Slice, r.Array:
				x.DataTypes.addType(f.Type.Elem())
				p.typeName = delphiTypeName(x.DataTypes.typesNames, f.Type.Elem())
				p.isArray = true
			case r.Struct:
				x.DataTypes.addType(f.Type)
				p.typeName = delphiTypeName(x.DataTypes.typesNames, f.Type)
				p.isStructure = true
			default:
				p.typeName = delphiPlainOldTypeName(f.Type)
			}

			m.params = append(m.params, p)
		}
	}

	returnType := met.Type.In(2).Elem()

	switch returnType.Kind() {
	case r.Slice:
		returnType = returnType.Elem()
		x.DataTypes.addType(returnType)
		m.retArray = true
		m.retPODType = len(delphiPlainOldTypeName(returnType)) > 0
		m.retDelphiType = delphiTypeName(x.DataTypes.typesNames, returnType)
	case r.Struct:
		if returnType.NumField() == 0 {
			m.procedure = true
		} else {
			x.DataTypes.addType(returnType)
			m.retDelphiType = delphiTypeName(x.DataTypes.typesNames, returnType)
			m.retPODType = len(delphiPlainOldTypeName(returnType)) > 0
		}
	default:
		m.retDelphiType = delphiPlainOldTypeName(returnType)
		m.retPODType = true
	}
	return
}

func (x method) remoteMethod(serviceName string) string {
	return fmt.Sprintf("'%s.%s'", serviceName, x.methodName)
}
