package data

import (
	"database/sql"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/elco"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func Open(createNew bool) (*sql.DB, error) {
	folderPath, err := elco.DataFolderPath()
	if err != nil {
		return nil, merry.Wrap(err)
	}
	fileName := filepath.Join(folderPath, "elco.sqlite")
	if createNew {
		if _, err := os.Stat(fileName); err == nil {
			if err := os.Remove(fileName); err != nil {
				return nil, merry.Appendf(err, "unable to delete database file: %s", fileName)
			}
		}
	}
	conn, err := sql.Open("sqlite3", fileName)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	conn.SetMaxIdleConns(1)
	conn.SetMaxOpenConns(1)
	conn.SetConnMaxLifetime(0)
	logrus.Infoln("open sqlite database:", fileName)
	return conn, nil
}
