package data

import (
	"database/sql"
	"github.com/fpawel/elco/internal/pkg/must"
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

func Close() {
	log.Info("close database")
	log.ErrIfFail(dbConn.Close)
}

func Open() {

	filename := filepath.Join(filepath.Dir(os.Args[0]), "elco.sqlite")

	dbConn = must.OpenSqliteDB(filename)
	if _, err := dbConn.Exec(SQLCreate); err != nil {
		panic(err)
	}
	if err := deleteEmptyParties(); err != nil {
		panic(err)
	}

	log.Debug("open database", "file", filename)
	DB = reform.NewDB(dbConn, sqlite3.Dialect, nil)
	DBx = sqlx.NewDb(dbConn, "sqlite3")

	_, _ = DBx.Exec(` ALTER TABLE party ADD max_d1 REAL DEFAULT NULL`)
	_, _ = DBx.Exec(` ALTER TABLE party ADD max_d2 REAL DEFAULT NULL`)
	_, _ = DBx.Exec(` ALTER TABLE party ADD max_d3 REAL DEFAULT NULL`)
	_, _ = DBx.Exec(` ALTER TABLE product_type ADD max_d1 REAL DEFAULT NULL`)
	_, _ = DBx.Exec(` ALTER TABLE product_type ADD max_d2 REAL DEFAULT NULL`)
	_, _ = DBx.Exec(` ALTER TABLE product_type ADD max_d3 REAL DEFAULT NULL`)
	_, _ = DBx.Exec(` ALTER TABLE product_type ADD k_sens20 REAL DEFAULT NULL`)
	_, _ = DBx.Exec(` ALTER TABLE product_type ADD points_method REAL DEFAULT NULL`)
	_, _ = DBx.Exec(` ALTER TABLE product_type ADD min_fon REAL DEFAULT NULL`)
	_, _ = DBx.Exec(` ALTER TABLE product_type ADD max_fon REAL DEFAULT NULL`)
	_, _ = DBx.Exec(` ALTER TABLE product_type ADD max_d_fon REAL DEFAULT NULL`)
	_, _ = DBx.Exec(` ALTER TABLE product_type ADD max_d_fon REAL DEFAULT NULL`)
	_, _ = DBx.Exec(` ALTER TABLE product_type ADD min_k_sens20 REAL DEFAULT NULL`)
	_, _ = DBx.Exec(` ALTER TABLE product_type ADD max_k_sens20 REAL DEFAULT NULL`)
	_, _ = DBx.Exec(` ALTER TABLE product_type ADD min_k_sens50 REAL DEFAULT NULL`)
	_, _ = DBx.Exec(` ALTER TABLE product_type ADD max_k_sens50 REAL DEFAULT NULL`)
	_, _ = DBx.Exec(` ALTER TABLE product_type ADD min_d_temp REAL DEFAULT NULL`)
	_, _ = DBx.Exec(` ALTER TABLE product_type ADD max_d_temp REAL DEFAULT NULL`)
	_, _ = DBx.Exec(` ALTER TABLE product_type ADD max_d_not_measured REAL DEFAULT NULL`)
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

var log = structlog.New()
