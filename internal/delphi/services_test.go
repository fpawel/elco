package delphi

import (
	"github.com/fpawel/elco/internal/svc"
	"os"
	r "reflect"
	"testing"
)

func TestRenderServices(t *testing.T) {
	types := []r.Type{
		r.TypeOf((*svc.PartiesCatalogue)(nil)),
		r.TypeOf((*svc.LastParty)(nil)),
		r.TypeOf((*svc.ProductTypes)(nil)),
	}
	RenderServices(types, map[string]string{
		"ProductInfo": "Product",
	}, os.Stdout)
}
