package pdf

import (
	"testing"
)

func TestPasportSou(t *testing.T) {
	if err := PasportSou(); err != nil {
		t.Error(err)
	}
}
