package main

import (
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

	dir := filepath.Join(os.Getenv("DELPHI_PATH"),
		"src", "github.com", "fpawel", "elcoui", "api")
	winapp.MustDir(dir)

	srvFn := filepath.Join(dir, "services.pas")
	srvFile, err := os.Open(srvFn)
	if os.IsNotExist(err) {
		srvFile, err = os.Create(srvFn)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer srvFile.Close()

	tpsFn := filepath.Join(dir, "server_data_types.pas")
	tpsFile, err := os.Open(tpsFn)
	if os.IsNotExist(err) {
		tpsFile, err = os.Create(tpsFn)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer tpsFile.Close()

	delphi.ServicesUnit(types, m, srvFile, tpsFile)
}
