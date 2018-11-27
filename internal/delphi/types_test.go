package delphi

import (
	"github.com/fpawel/elco/internal/svc"
	"os"
	r "reflect"
	"testing"
)

func TestPascalSource(t *testing.T) {
	x := NewTypes([]r.Type{
		r.TypeOf((*svc.Party)(nil)).Elem(),
	}, map[string]string{
		"ProductInfo": "Product",
	})
	x.WriteUnit(os.Stdout)
}
