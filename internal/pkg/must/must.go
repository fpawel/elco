package must

import (
	"database/sql"
	"github.com/fpawel/elco/internal/pkg"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"os"
	"syscall"
)

// AbortIf should point to FatalIf or PanicIf or similar user-provided
// function which will interrupt execution in case it's param is not nil.
var AbortIf = PanicIf

// Alternative names to consider:
//  must.OK()
//  must.BeNil()
//  must.OrPanic()
//  must.OrAbort()
//  must.OrDie()

// NoErr is just a synonym for AbortIf.
func NoErr(err error) {
	AbortIf(err)
}

// PanicIf will call panic(err) in case given err is not nil.
func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

// Close is a wrapper for os.File.Close, …
func Close(f io.Closer) {
	err := f.Close()
	AbortIf(err)
}

// Create is a wrapper for os.Create.
func Create(name string) *os.File {
	f, err := os.Create(name)
	AbortIf(err)
	return f
}

// Decoder is an interface compatible with json.Decoder, gob.Decoder,
// xml.Decoder, …
type Decoder interface {
	Decode(v interface{}) error
}

// Encoder is an interface compatible with json.Encoder, gob.Encoder,
// xml.Encoder, …
type Encoder interface {
	Encode(v interface{}) error
}

func UTF16PtrFromString(s string) *uint16 {
	p, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}
	return p
}

func OpenSqliteDB(fileName string) *sql.DB {
	conn, err := pkg.OpenSqliteDB(fileName)
	NoErr(err)
	return conn
}

func EnsureDir(dir string) {
	AbortIf(pkg.EnsureDir(dir))
}
