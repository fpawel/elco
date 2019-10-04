package main

import (
	"fmt"
	"log"
	"path"
	r "reflect"
	"strings"
)

type NotifyServicesSrc struct {
	delphiHandlersTypes map[string]string
	goImports           map[string]struct{}
	services            []notifyService
	DataTypes           *TypesUnit
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

func NewNotifyServicesSrc(d *TypesUnit, services []NotifyServiceType) *NotifyServicesSrc {
	x := &NotifyServicesSrc{
		DataTypes:           d,
		delphiHandlersTypes: make(map[string]string),
		goImports:           make(map[string]struct{}),
	}
	for _, s := range services {
		if s.ParamType.Kind() == r.Struct && s.ParamType.NumField() == 0 {
			x.services = append(x.services, notifyService{
				serviceName: s.ServiceName,
				handlerType: "TProcedure",
			})
			x.delphiHandlersTypes["TProcedure"] = "reference to procedure"
			continue
		}
		t, err := x.DataTypes.add(s.ParamType)
		if err != nil {
			log.Fatalln("notify_service:", s.ServiceName, "error:", err)
		}
		handlerTypeName := strings.Title(t.TypeName() + "Handler")
		if handlerTypeName[0] != 'T' {
			handlerTypeName = "T" + handlerTypeName
		}
		y := notifyService{
			serviceName: s.ServiceName,
			typeName:    t.TypeName(),
			handlerType: handlerTypeName,
			goType:      s.ParamType.Name(),
		}
		if len(s.ParamType.PkgPath()) > 0 {
			x.goImports[s.ParamType.PkgPath()] = struct{}{}
			y.goType = path.Base(s.ParamType.PkgPath()) + "." + y.goType

		}
		switch s.ParamType.Kind() {

		case r.String:
			y.instructionGetFromStr = "str"
			y.notifyFunc = "w.SendString"
			y.instructionArg = "arg"

		case r.Int,
			r.Int8, r.Int16, r.Int32, r.Int64,
			r.Uint8, r.Uint16, r.Uint32, r.Uint64:
			y.instructionGetFromStr = "StrToInt(str)"
			y.notifyFunc = "w.SendString"
			y.instructionArg = "fmt.Sprintf(\"%d\", arg)"

		case r.Float32, r.Float64:
			y.instructionGetFromStr = "str_to_float(str)"
			y.notifyFunc = "w.SendString"
			y.instructionArg = "fmt.Sprintf(\"%v\", arg)"

		case r.Bool:
			y.instructionGetFromStr = "StrToBool(str)"
			y.notifyFunc = "w.SendString"
			y.instructionArg = "fmt.Sprintf(\"%v\", arg)"

		case r.Struct:
			y.instructionGetFromStr = fmt.Sprintf("_deserializer.deserialize<%s>(str)", t.TypeName())
			y.notifyFunc = "w.SendJson"
			y.instructionArg = "arg"
		default:
			panic(fmt.Sprintf("wrong type %q: %v", s.ServiceName, s.ParamType))
		}

		x.delphiHandlersTypes[y.handlerType] = fmt.Sprintf("reference to procedure (x:%s)", y.typeName)
		x.services = append(x.services, y)
	}
	return x
}
