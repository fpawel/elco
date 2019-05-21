package data

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/sqlite3"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"
)

func TestConcurrent(t *testing.T) {
	file, err := ioutil.TempFile("", "db_*.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	_ = file.Close()
	defer func() {
		_ = os.Remove(file.Name())
	}()

	dbConn, err := sql.Open("sqlite3", file.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = dbConn.Close()
	}()
	dbConn.SetMaxIdleConns(1)
	dbConn.SetMaxOpenConns(1)
	dbConn.SetConnMaxLifetime(0)

	db, err := Open(false)
	if err != nil {
		t.Fatal(err)
	}
	dbx := sqlx.NewDb(dbConn, "sqlite3")

	wg := sync.WaitGroup{}
	wg.Add(200)
	for i := 0; i < 100; i++ {
		i := i
		go func() {
			_, err := createNewPartyReform(reform.NewDB(db, sqlite3.Dialect, nil))
			if err != nil {
				t.Errorf("%d: %v", i, err)
			}
			wg.Done()
		}()
		go func() {
			_, err := createNewPartySqlx(dbx)
			if err != nil {
				t.Errorf("%d: %v", i, err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func createNewPartyReform(db *reform.DB) (int64, error) {
	r, err := db.Exec(`INSERT INTO party DEFAULT VALUES`)
	if err != nil {
		return 0, err
	}
	partyID, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}
	if r, err = db.Exec(`INSERT INTO product(party_id, serial, place) VALUES (?, 1, 0)`, partyID); err != nil {
		return 0, err
	}
	log.Println("reform: new party created:", partyID)
	return partyID, nil
}

func createNewPartySqlx(db *sqlx.DB) (int64, error) {
	r, err := db.Exec(`INSERT INTO party DEFAULT VALUES`)
	if err != nil {
		return 0, err
	}
	partyID, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}
	if r, err = db.Exec(`INSERT INTO product(party_id, serial, place) VALUES (?, 1, 0)`, partyID); err != nil {
		return 0, err
	}
	log.Println("sqlx: new party created:", partyID)
	return partyID, nil
}
