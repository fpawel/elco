package delphi

import (
	"github.com/fpawel/elco/internal/svc"
	"os"
	r "reflect"
	"testing"
)

func TestPascalSource(t *testing.T) {
	WriteDataTypesUnit(
		os.Stdout,
		NewDataTypes([]r.Type{
			r.TypeOf((*svc.Party)(nil)).Elem(),
		}, map[string]string{
			"ProductInfo": "Product",
		}),
	)
}
