package db

import (
	"database/sql"
	"fmt"
	"github.com/fpawel/elco/internal/app"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/sqlite3"
	"log"
	"os"
	"testing"
)

func TestDB(t *testing.T) {
	// Get *sql.DB as usual. Sqlite3 example:
	conn, err := sql.Open("sqlite3", app.DataFileName())
	if err != nil {
		log.Fatal(err)
	}

	// Use new *log.Logger for logging.
	logger := log.New(os.Stderr, "SQL: ", log.Flags())

	// Create *reform.DB instance with simple logger.
	// Any Printf-like function (fmt.Printf, log.Printf, testing.T.Logf, etc) can be used with NewPrintfLogger.
	// Change dialect for other databases.
	db := reform.NewDB(conn, sqlite3.Dialect, reform.NewPrintfLogger(logger.Printf))

	var lastParty LastParty
	if err := db.SelectOneTo(&lastParty, ""); err != nil {
		log.Fatal(err)
	}
	fmt.Println(lastParty)

	products, err := db.SelectAllFrom(ProductInfoTable, "WHERE party_id = ? ORDER BY place", 4)
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range products {
		fmt.Println(p)
	}

}
