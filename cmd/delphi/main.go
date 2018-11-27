package main

import (
	"github.com/fpawel/elco/internal/daemon"
	"github.com/fpawel/elco/internal/delphi"
	"github.com/fpawel/elco/internal/svc"
	"github.com/fpawel/goutils/winapp"
	"log"
	"os"
	"path/filepath"
	r "reflect"
)

func main() {
	types := []r.Type{
		r.TypeOf((*svc.PartiesCatalogue)(nil)),
		r.TypeOf((*svc.LastParty)(nil)),
		r.TypeOf((*svc.ProductTypes)(nil)),
	}
	m := map[string]string{
		"ProductInfo": "Product",
	}

	dir := filepath.Join(os.Getenv("DELPHIPATH"),
		"src", "github.com", "fpawel", "elcoui", "api")
	winapp.MustDir(dir)

	srvFile, err := os.Create(filepath.Join(dir, "services.pas"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = srvFile.Close()
	}()

	fn := filepath.Join(dir, "server_data_types.pas")
	tpsFile, err := os.Create(fn)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = tpsFile.Close()
	}()
	delphi.ServicesUnit(daemon.PipeName, types, m, srvFile, tpsFile)
}
