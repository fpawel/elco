package main

import (
	"github.com/fpawel/elco/internal/api"
	"os"
	r "reflect"
	"testing"
)

func TestServices(t *testing.T) {
	types := []r.Type{
		r.TypeOf((*api.PartiesCatalogue)(nil)),
		r.TypeOf((*api.LastParty)(nil)),
		r.TypeOf((*api.ProductTypes)(nil)),
	}
	ServicesUnit(types, map[string]string{
		"ProductInfo": "Product",
	}, os.Stdout)
}
