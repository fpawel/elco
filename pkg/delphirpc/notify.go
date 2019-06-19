package delphirpc

import (
	"fmt"
	"path"
	r "reflect"
)

type NotifyServicesSrc struct {
	unitName string
	interfaceUses,
	implUses []string
	types     map[string]string
	goImports map[string]struct{}
	services  []notifyService
	DataTypes *TypesSrc
}

type NotifyServiceType struct {
	ServiceName string
	ParamType   r.Type
}

type notifyService struct {
	serviceName,
	typeName,
	handlerType,
	notifyFunc,
	goType,
	instructionGetFromStr,
	instructionArg string
}

func NewNotifyServicesSrc(unitName string, d *TypesSrc, services []NotifyServiceType) *NotifyServicesSrc {
	x := &NotifyServicesSrc{
		unitName:      unitName,
		interfaceUses: []string{d.unitName, "superobject", "Winapi.Windows", "Winapi.Messages"},
		implUses:      []string{"Grijjy.Bson.Serialization", "stringutils", "sysutils"},
		DataTypes:     d,
		types:         make(map[string]string),
		goImports:     make(map[string]struct{}),
	}
	for _, s := range services {
		x.DataTypes.addType(s.ParamType)

		t := delphiTypeName(x.DataTypes.typesNames, s.ParamType)
		y := notifyService{
			serviceName: s.ServiceName,
			typeName:    t,
			handlerType: strEnsureFirstT(t) + "Handler",
			goType:      s.ParamType.Name(),
		}
		x.types[y.typeName] = y.handlerType

		if len(s.ParamType.PkgPath()) > 0 {
			x.goImports[s.ParamType.PkgPath()] = struct{}{}
			y.goType = path.Base(s.ParamType.PkgPath()) + "." + y.goType

		}

		switch s.ParamType.Kind() {

		case r.String:
			y.instructionGetFromStr = "str"
			y.notifyFunc = "NotifyStr"
			y.instructionArg = "arg"

		case r.Int,
			r.Int8, r.Int16, r.Int32, r.Int64,
			r.Uint8, r.Uint16, r.Uint32, r.Uint64:
			y.instructionGetFromStr = "StrToInt(str)"
			y.notifyFunc = "NotifyStr"
			y.instructionArg = "fmt.Sprintf(\"%d\", arg)"

		case r.Float32, r.Float64:
			y.instructionGetFromStr = "str_to_float(str)"
			y.notifyFunc = "NotifyStr"
			y.instructionArg = "fmt.Sprintf(\"%v\", arg)"

		case r.Bool:
			y.instructionGetFromStr = "StrToBool(str)"
			y.notifyFunc = "NotifyStr"
			y.instructionArg = "fmt.Sprintf(\"%v\", arg)"

		case r.Struct:
			y.instructionGetFromStr = fmt.Sprintf("_deserializer.deserialize<%s>(str)", t)
			y.notifyFunc = "NotifyJson"
			y.instructionArg = "arg"

		default:
			panic(fmt.Sprintf("wrong type %q: %v", s.ServiceName, s.ParamType))
		}

		x.services = append(x.services, y)
	}
	return x
}
