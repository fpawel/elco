package pdf

import (
	"github.com/fpawel/elco/internal/data"
	"github.com/powerman/must"
	"testing"
)

func TestPdf(t *testing.T) {
	data.Open()
	must.AbortIf(RunPartyID(data.GetLastPartyID()))
}
