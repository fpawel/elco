package main

import (
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/daemon"
	"github.com/fpawel/goutils/winapp"
	"log"
	"os"
	"path/filepath"
	r "reflect"
)

func main() {
	types := []r.Type{
		r.TypeOf((*api.PartiesCatalogue)(nil)),
		r.TypeOf((*api.LastParty)(nil)),
		r.TypeOf((*api.ProductTypes)(nil)),
		r.TypeOf((*api.ProductFirmware)(nil)),
		r.TypeOf((*api.SettingsSvc)(nil)),
	}
	m := map[string]string{
		"ProductInfo": "Product",
	}

	dir := filepath.Join(os.Getenv("DELPHIPATH"),
		"src", "github.com", "fpawel", "elcoui", "api")
	winapp.MustDir(dir)

	openFile := func(fileName string) *os.File {
		file, err := os.Create(filepath.Join(dir, fileName))
		if err != nil {
			log.Fatal(err)
		}
		return file
	}

	servicesSrc := NewServicesSrc(daemon.PipeName, types, m)

	notifySvcSrc := NewNotifyServicesSrc(servicesSrc.dataTypes, []notifyServiceType{
		{
			serviceName: "MessageBox",
			paramType:   r.TypeOf((*api.TextMessage)(nil)).Elem(),
		},
		{
			serviceName: "StatusMessage",
			paramType:   r.TypeOf((*string)(nil)).Elem(),
		},
	})

	file := openFile("services.pas")
	servicesSrc.WriteUnit(file)
	file.Close()

	file = openFile("server_data_types.pas")
	servicesSrc.dataTypes.WriteUnit(file)
	file.Close()

	dir = filepath.Join(os.Getenv("GOPATH"),
		"src", "github.com", "fpawel", "elco", "notify")
	winapp.MustDir(dir)

	file = openFile("notify_services.pas")
	notifySvcSrc.WriteUnit(file)
	file.Close()

	dir = filepath.Join(os.Getenv("GOPATH"),
		"src", "github.com", "fpawel", "elco", "internal", "api", "notify")
	file, err := os.Create(filepath.Join(dir, "api_generated.go"))
	if err != nil {
		log.Fatal(err)
	}
	notifySvcSrc.WriteGoFile(file)
	file.Close()

}
