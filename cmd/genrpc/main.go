package main

import (
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/elco/pkg/delphirpc"
	"github.com/fpawel/elco/pkg/winapp"
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
		r.TypeOf((*api.PlaceFirmware)(nil)),
		r.TypeOf((*api.SettingsSvc)(nil)),
		r.TypeOf((*api.RunnerSvc)(nil)),
	}
	m := map[string]string{
		"ProductInfo": "Product",
		"WorkInfo":    "JournalWork",
		"EntryInfo":   "JournalEntry",
	}

	dir := filepath.Join(os.Getenv("DELPHIPATH"),
		"src", "github.com", "fpawel", "elco-gui", "api")
	winapp.EnsuredDirectory(dir)

	openFile := func(fileName string) *os.File {
		file, err := os.Create(filepath.Join(dir, fileName))
		if err != nil {
			log.Fatal(err)
		}
		return file
	}

	servicesSrc := delphirpc.NewServicesSrc(elco.PipeName, "services", "server_data_types", types, m)

	notifySvcSrc := delphirpc.NewNotifyServicesSrc("notify_services", servicesSrc.DataTypes, []delphirpc.NotifyServiceType{
		{
			"ReadCurrent",
			r.TypeOf((*api.ReadCurrent)(nil)).Elem(),
		},
		{
			"ErrorOccurred",
			r.TypeOf((*string)(nil)).Elem(),
		},
		{
			"WorkComplete",
			r.TypeOf((*string)(nil)).Elem(),
		},
		{
			"WorkStarted",
			r.TypeOf((*string)(nil)).Elem(),
		},
		{
			"WorkStopped",
			r.TypeOf((*string)(nil)).Elem(),
		},
		{
			"Status",
			r.TypeOf((*string)(nil)).Elem(),
		},

		{
			"TraceTemperatureInfo",
			r.TypeOf((*string)(nil)).Elem(),
		},

		{
			"TraceTemperatureError",
			r.TypeOf((*string)(nil)).Elem(),
		},

		{
			"Warning",
			r.TypeOf((*string)(nil)).Elem(),
		},
		{
			"Delay",
			r.TypeOf((*api.DelayInfo)(nil)).Elem(),
		},
		{
			"LastPartyChanged",
			r.TypeOf((*data.Party)(nil)).Elem(),
		},
		{
			"StartServerApplication",
			r.TypeOf((*string)(nil)).Elem(),
		},
		{
			"ReadFirmware",
			r.TypeOf((*data.FirmwareInfo)(nil)).Elem(),
		},

		{
			"Panic",
			r.TypeOf((*string)(nil)).Elem(),
		},

		{
			"WriteConsole",
			r.TypeOf((*string)(nil)).Elem(),
		},
	})

	file := openFile("services.pas")
	servicesSrc.WriteUnit(file)
	if err := file.Close(); err != nil {
		panic(err)
	}

	file = openFile("server_data_types.pas")
	servicesSrc.DataTypes.WriteUnit(file)
	if err := file.Close(); err != nil {
		panic(err)
	}

	file = openFile("notify_services.pas")
	notifySvcSrc.WriteUnit(file)
	if err := file.Close(); err != nil {
		panic(err)
	}

	dir = filepath.Join(os.Getenv("GOPATH"),
		"src", "github.com", "fpawel", "elco", "internal", "api", "notify")
	file, err := os.Create(filepath.Join(dir, "api_generated.go"))
	if err != nil {
		log.Fatal(err)
	}
	notifySvcSrc.WriteGoFile(file)
	if err := file.Close(); err != nil {
		panic(err)
	}

}
