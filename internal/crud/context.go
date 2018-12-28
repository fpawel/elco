package crud

import (
	"github.com/fpawel/elco/internal/app"
	"github.com/fpawel/elco/internal/data"
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

func NewDBContext(logger reform.Logger) DBContext {
	dbx := dbutils.MustOpen(app.DataFileName(), "sqlite3")
	dbx.MustExec(data.SQLCreate)
	data.DeleteEmptyRecords(dbx)
	data.EnsureParty(dbx)
	dbr := reform.NewDB(dbx.DB, sqlite3.Dialect, logger)

	return DBContext{
		dbContext{
			dbx: dbx,
			dbr: dbr,
			mu:  new(sync.Mutex),
		},
	}
}
func (x DBContext) Close() error {
	return x.dbx.Close()
}

func (x DBContext) PartiesCatalogue() PartiesCatalogue {
	return PartiesCatalogue{dbContext: x.dbContext}
}

func (x DBContext) LastParty() LastParty {
	return LastParty{dbContext: x.dbContext}
}

func (x DBContext) ProductTypes() ProductTypes {
	return ProductTypes{dbContext: x.dbContext}
}

func (x DBContext) ProductFirmware() ProductFirmware {
	return ProductFirmware{dbContext: x.dbContext}
}

func (x dbContext) ListProductTypesNames() (names []string) {
	x.mu.Lock()
	defer x.mu.Unlock()
	for _, p := range data.ListProductTypes(x.dbr) {
		names = append(names, p.ProductTypeName)
	}
	return
}

func (x dbContext) ListGases() []data.Gas {
	x.mu.Lock()
	defer x.mu.Unlock()
	return data.ListGases(x.dbr)
}

func (x dbContext) ListUnits() []data.Units {
	x.mu.Lock()
	defer x.mu.Unlock()
	return data.ListUnits(x.dbr)
}

func (x dbContext) SaveProduct(p *data.Product) {
	x.mu.Lock()
	defer x.mu.Unlock()
	if err := x.dbr.Save(p); err != nil {
		panic(err)
	}
}

func (x dbContext) NewParty() data.Party {
	x.mu.Lock()
	defer x.mu.Unlock()
	data.CreateNewParty(x.dbx)
	return data.MustLastParty(x.dbr)
}
