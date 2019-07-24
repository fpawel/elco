package main

import (
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/delphirpc"
	"github.com/fpawel/gohelp/winapp"
	"log"
	"os"
	"path/filepath"
	r "reflect"
)

func main() {
	types := []r.Type{
		r.TypeOf((*api.PartiesCatalogueSvc)(nil)),
		r.TypeOf((*api.LastPartySvc)(nil)),
		r.TypeOf((*api.ProductTypesSvc)(nil)),
		r.TypeOf((*api.PlaceFirmware)(nil)),
		r.TypeOf((*api.SettingsSvc)(nil)),
		r.TypeOf((*api.RunnerSvc)(nil)),
		r.TypeOf((*api.PdfSvc)(nil)),
		r.TypeOf((*api.PeerSvc)(nil)),
	}
	m := map[string]string{
		"ProductInfo": "Product",
		"WorkInfo":    "JournalWork",
		"EntryInfo":   "JournalEntry",
	}

	dir := filepath.Join(os.Getenv("DELPHIPATH"),
		"src", "github.com", "fpawel", "elco-gui", "api")
	_ = winapp.EnsuredDirectory(dir)

	openFile := func(fileName string) *os.File {
		file, err := os.Create(filepath.Join(dir, fileName))
		if err != nil {
			log.Fatal(err)
		}
		return file
	}

	servicesSrc := delphirpc.NewServicesSrc("services", "server_data_types", types, m)

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
			"Ktx500Info",
			r.TypeOf((*api.Ktx500Info)(nil)).Elem(),
		},

		{
			"Ktx500Error",
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
			"EndDelay",
			r.TypeOf((*string)(nil)).Elem(),
		},
		{
			"LastPartyChanged",
			r.TypeOf((*data.Party)(nil)).Elem(),
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
		{
			"ReadPlace",
			r.TypeOf((*int)(nil)).Elem(),
		},
		{
			"ReadBlock",
			r.TypeOf((*int)(nil)).Elem(),
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
