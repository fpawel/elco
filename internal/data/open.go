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


	if _, err = conn.Exec(SQLCreate); err != nil {
		return nil, merry.Wrap(err)
	}
	if _, err = conn.Exec(`
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
`); err != nil {
		return nil, merry.Wrap(err)
	}
	logrus.Infoln("open sqlite database:", fileName)
	return conn, nil
}
