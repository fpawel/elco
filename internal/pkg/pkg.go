package pkg

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

func OpenSqliteDB(fileName string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", fileName)
	if err != nil {
		return nil, err
	}
	conn.SetMaxIdleConns(1)
	conn.SetMaxOpenConns(1)
	conn.SetConnMaxLifetime(0)
	return conn, err
}

func EnsureDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) { // создать каталог если его нет
		err = os.MkdirAll(dir, os.ModePerm)
	}
	return err
}
