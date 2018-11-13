package crud

import (
	"github.com/fpawel/elco/internal/app"
	"github.com/fpawel/goutils/dbutils"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/sqlite3"
	"sync"
)

type Context struct {
	mu   sync.Mutex
	conn *sqlx.DB
	dbr  *reform.DB
}

func NewContext(logger reform.Logger) *Context {
	conn := dbutils.MustOpen(app.DataFileName(), "sqlite3")
	return &Context{
		conn: conn,
		dbr:  reform.NewDB(conn.DB, sqlite3.Dialect, logger),
	}
}
func (x *Context) Close() error {
	return x.conn.Close()
}

func (x *Context) PartiesCatalogue() PartiesCatalogue {
	return PartiesCatalogue{
		dbr:  x.dbr,
		conn: x.conn,
		mu:   &x.mu,
	}
}
