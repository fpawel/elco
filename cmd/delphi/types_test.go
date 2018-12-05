package main

import (
	"github.com/fpawel/elco/internal/api"
	"os"
	r "reflect"
	"testing"
)

func TestPascalSource(t *testing.T) {
	x := NewTypes([]r.Type{
		r.TypeOf((*api.Party)(nil)).Elem(),
	}, map[string]string{
		"ProductInfo": "Product",
	})
	x.WriteUnit(os.Stdout)
}
