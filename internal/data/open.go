package data

import (
	"database/sql"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/sqlite3"
)

func Open(dbConn *sql.DB, logger reform.Logger) (*reform.DB, error) {
	db := reform.NewDB(dbConn, sqlite3.Dialect, logger)
	_, err := db.Exec(SQLCreate)
	if err != nil {
		return nil, err
	}
	return db, DeleteEmptyRecords(db)
}
