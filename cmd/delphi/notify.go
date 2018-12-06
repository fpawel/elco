package main

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
	dataTypes *TypesSrc
}

type notifyServiceType struct {
	serviceName string
	paramType   r.Type
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

func NewNotifyServicesSrc(d *TypesSrc, services []notifyServiceType) *NotifyServicesSrc {
	x := &NotifyServicesSrc{
		unitName:      "notify_services",
		interfaceUses: []string{"server_data_types", "superobject", "Winapi.Windows", "Winapi.Messages"},
		implUses:      []string{"Rest.Json", "stringutils", "sysutils"},
		dataTypes:     d,
		types:         make(map[string]string),
		goImports:     make(map[string]struct{}),
	}
	for _, s := range services {
		x.dataTypes.addType(s.paramType)

		t := delphiTypeName(x.dataTypes.typesNames, s.paramType)
		y := notifyService{
			serviceName: s.serviceName,
			typeName:    t,
			handlerType: strEnsureFirstT(t) + "Handler",
			goType:      s.paramType.Name(),
		}
		x.types[y.typeName] = y.handlerType

		if len(s.paramType.PkgPath()) > 0 {
			x.goImports[s.paramType.PkgPath()] = struct{}{}
			y.goType = path.Base(s.paramType.PkgPath()) + "." + y.goType

		}

		switch s.paramType.Kind() {

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
			y.instructionGetFromStr = fmt.Sprintf(
				"TJson.JsonToObject<%s>(str)", t)
			y.notifyFunc = "NotifyJson"
			y.instructionArg = "arg"

		default:
			panic(fmt.Sprintf("wrong type %q: %v", s.serviceName, s.paramType))
		}

		x.services = append(x.services, y)
	}
	return x
}
