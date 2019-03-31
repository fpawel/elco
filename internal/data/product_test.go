package data

import (
	"fmt"
	"reflect"
	"testing"
)

func TestPrintProductTypeInfo(t *testing.T) {
	p := ProductInfo{}
	ti := reflect.TypeOf(p)
	for i := 0; i < ti.NumField(); i++ {
		field := ti.Field(i)
		fmt.Printf("F%s := FieldValues['%s'];\n", field.Name, field.Tag.Get("reform"))
	}
}
