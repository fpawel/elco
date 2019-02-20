package journal

import (
	"database/sql"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/sqlite3"
)

func Open(dbConn *sql.DB, logger reform.Logger) (*reform.DB, error) {
	db := reform.NewDB(dbConn, sqlite3.Dialect, logger)
	_, err := db.Exec(SQLCreate)
	return db, err
}

func (s Entry) EntryInfo(workName string) EntryInfo {
	return EntryInfo{
		CreatedAt: s.CreatedAt,
		EntryID:   s.EntryID,
		WorkID:    s.WorkID,
		Message:   s.Message,
		Level:     s.Level,
		File:      s.File,
		Line:      s.Line,
		Stack:     s.Stack,
		WorkName:  workName,
	}
}
