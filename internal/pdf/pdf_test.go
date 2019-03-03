package pdf

import (
	"github.com/fpawel/elco/internal/assets"
	"github.com/fpawel/elco/internal/data"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/sqlite3"
	"testing"
)

func TestPdf(t *testing.T) {
	dbConn, err := data.Open(false)
	if err != nil {
		t.Error(err)
	}
	if err := Run(reform.NewDB(dbConn, sqlite3.Dialect, nil)); err != nil {
		t.Error(err)
	}
}

func init() {
	assets.Ensure()
}
