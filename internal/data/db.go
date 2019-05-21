package data

import (
	"database/sql"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/pkg/winapp"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/powerman/structlog"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/sqlite3"
	"os"
	"path/filepath"
)

var (
	DB     *reform.DB
	DBx    *sqlx.DB
	dbConn *sql.DB
)

func Close() error {
	return dbConn.Close()
}

func Open(createNew bool) {
	dir, err := Dir()
	if err != nil {
		panic(err)
	}
	fileName := filepath.Join(dir, "elco.sqlite")
	if createNew {
		if _, err := os.Stat(fileName); err == nil {
			if err := os.Remove(fileName); err != nil {
				panic(merry.Appendf(err, "unable to delete database file: %s", fileName))
			}
		}
	}
	dbConn, err = sql.Open("sqlite3", fileName)
	if err != nil {
		panic(err)
	}
	dbConn.SetMaxIdleConns(1)
	dbConn.SetMaxOpenConns(1)
	dbConn.SetConnMaxLifetime(0)

	if _, err = dbConn.Exec(SQLCreate); err != nil {
		panic(err)
	}
	if err = deleteEmptyParties(); err != nil {
		panic(err)
	}
	structlog.New().Info("open sqlite db", "file", fileName)
	DB = reform.NewDB(dbConn, sqlite3.Dialect, nil)
	DBx = sqlx.NewDb(dbConn, "sqlite3")
}

func Dir() (string, error) {
	dir, err := winapp.AppDataFolderPath()
	if err != nil {
		return "", merry.Wrap(err)
	}
	dir = filepath.Join(dir, "elco")
	err = winapp.EnsuredDirectory(dir)
	if err != nil {
		return "", merry.Wrap(err)
	}
	return dir, nil
}

func deleteEmptyParties() error {
	_, err := dbConn.Exec(`
DELETE
FROM product
WHERE party_id NOT IN (SELECT party_id FROM last_party)  AND
  serial ISNULL  AND
  (product_type_name ISNULL OR LENGTH(trim(product_type_name)) = 0)  AND
  (note ISNULL OR LENGTH(trim(note)) = 0)  AND
  i_f_minus20 ISNULL  AND
  i_f_plus20 ISNULL  AND
  i_f_plus50 ISNULL  AND
  i_s_minus20 ISNULL  AND
  i_s_plus20 ISNULL  AND
  i_s_plus50 ISNULL  AND
  i13 ISNULL  AND
  i24 ISNULL  AND
  i35 ISNULL  AND
  i26 ISNULL  AND
  i17 ISNULL  AND
  not_measured ISNULL  AND
  (firmware ISNULL OR LENGTH(firmware) = 0)  AND
  old_product_id ISNULL  AND
  old_serial ISNULL;
DELETE
FROM party
WHERE NOT EXISTS(SELECT product_id FROM product WHERE party.party_id = product.party_id);
`)
	return err
}
