package main

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

type delphiServiceMethod struct {
	methodName  string
	params      []param
	ret         *delphiType
	namedParams bool
}

type param struct {
	name       string
	delphiType delphiType
}

func (x *delphiServiceMethod) addParam(paramName string, paramType reflect.Type, typesUnit *TypesUnit) {
	t, err := typesUnit.add(paramType)
	if err != nil {
		log.Fatalln("method:", x.methodName, "param:", paramName, "error:", err)
	}
	x.params = append(x.params, param{
		name:       paramName,
		delphiType: t,
	})
	fmt.Println("\tparam:", paramName, ": ", t.TypeName())
}

func (x delphiServiceMethod) decl(prefix string) string {
	s := "class "
	if x.ret != nil {
		s += "function"
	} else {
		s += "procedure"
	}
	s += fmt.Sprintf(" %s%s", prefix, x.methodName)
	if len(x.params) > 0 {
		xs := make([]string, len(x.params))
		for i, p := range x.params {
			xs[i] = fmt.Sprintf("%s:%s", p.name, p.delphiType.TypeName())
		}
		s += "(" + strings.Join(xs, "; ") + ")"
	}

	if x.ret != nil {
		s += ":" + x.ret.TypeName()
	}
	return s
}

func setParamFieldInstruction(paramTypeKind delphiTypeKind, paramName string) string {
	if paramTypeKind == delphiRecord {
		return fmt.Sprintf("TgoBsonSerializer.serialize(%s, s); req['%s'] := SO(s)", paramName, paramName)
	}
	return fmt.Sprintf("SuperObject_SetField(req, '%s', %s)", paramName, paramName)
}
