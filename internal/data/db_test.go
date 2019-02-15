package data

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"sync"
	"testing"
)

func TestConcurrent(t *testing.T) {
	f, err := os.Create(`d:\test.db`)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	dbConn, err := sql.Open("sqlite3", `d:\test.db`)
	if err != nil {
		t.Fatal(err)
	}
	defer dbConn.Close()
	dbConn.SetMaxIdleConns(1)
	dbConn.SetMaxOpenConns(1)
	dbConn.SetConnMaxLifetime(0)

	db, err := Open(dbConn, nil)
	if err != nil {
		t.Fatal(err)
	}
	const count = 100
	wg := sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < count; i++ {
		i := i
		go func() {
			_, err := CreateNewParty(db)
			if err != nil {
				t.Errorf("%d: %v", i, err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
