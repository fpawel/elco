package main

import (
	"fmt"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/pkg/must"
	"os"
	"path/filepath"
	r "reflect"
)

func main() {
	delphiSrcDir := filepath.Join(os.Getenv("DELPHIPATH"), "src", "github.com", "fpawel", "elco-gui", "api")

	must.EnsureDir(delphiSrcDir)

	servicesSrc := NewServicesUnit("elco", []r.Type{
		r.TypeOf((*api.PartiesCatalogueSvc)(nil)),
		r.TypeOf((*api.LastPartySvc)(nil)),
		r.TypeOf((*api.ProductTypesSvc)(nil)),
		r.TypeOf((*api.PlaceFirmware)(nil)),
		r.TypeOf((*api.RunnerSvc)(nil)),
		r.TypeOf((*api.PdfSvc)(nil)),
		r.TypeOf((*api.ConfigSvc)(nil)),
		r.TypeOf((*api.ProductsCatalogueSvc)(nil)),
	})
	notifySvcSrc := NewNotifyServicesSrc(servicesSrc.TypesUnit, []NotifyServiceType{
		{
			"ReadCurrent",
			r.TypeOf((*api.ReadCurrent)(nil)).Elem(),
		},
		{
			"WorkComplete",
			r.TypeOf((*api.WorkResult)(nil)).Elem(),
		},
		{
			"WorkStarted",
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
			r.TypeOf((*api.Party1)(nil)).Elem(),
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
		{
			"ScriptLine",
			r.TypeOf((*api.ScriptLine)(nil)).Elem(),
		},
	})

	createFile := func(fileName string) *os.File {
		fileName = filepath.Join(delphiSrcDir, fileName)
		fmt.Println("file:", fileName)
		return must.Create(fileName)
	}

	file := createFile("services.pas")
	servicesSrc.WriteUnit(file)
	must.Close(file)

	file = createFile("server_data_types.pas")
	servicesSrc.TypesUnit.WriteUnit(file)
	must.Close(file)

	file = createFile("notify_services.pas")
	notifySvcSrc.WriteUnit(file)
	must.Close(file)

	file = must.Create(filepath.Join(os.Getenv("GOPATH"),
		"src", "github.com", "fpawel", "elco", "internal", "api", "notify", "api_generated.go"))
	notifySvcSrc.WriteGoFile(file)
	must.Close(file)
}
