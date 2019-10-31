package pdf

import (
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/pkg/must"
	"testing"
)

func TestPdf(t *testing.T) {
	data.Open()
	must.AbortIf(RunPartyID(data.LastPartyID()))
}
