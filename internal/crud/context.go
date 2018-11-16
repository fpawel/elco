package crud

import (
	"github.com/fpawel/elco/internal/app"
	"github.com/fpawel/elco/internal/crud/data"
	"github.com/fpawel/goutils/dbutils"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/sqlite3"
	"sync"
)

type DBContext struct {
	dbContext
}

type dbContext struct {
	mu  *sync.Mutex
	dbx *sqlx.DB
	dbr *reform.DB
}

func NewDBContext(logger reform.Logger) *DBContext {
	dbx := dbutils.MustOpen(app.DataFileName(), "sqlite3")
	data.DeleteEmptyRecords(dbx)
	return &DBContext{
		dbContext{
			dbx: dbx,
			dbr: reform.NewDB(dbx.DB, sqlite3.Dialect, logger),
			mu:  new(sync.Mutex),
		},
	}
}
func (x *DBContext) Close() error {
	return x.dbx.Close()
}

func (x *DBContext) PartiesCatalogue() PartiesCatalogue {
	return PartiesCatalogue{dbContext: x.dbContext}
}

func (x *DBContext) LastParty() LastParty {
	return LastParty{dbContext: x.dbContext}
}
