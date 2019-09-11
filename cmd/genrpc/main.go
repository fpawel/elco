package main

import (
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/gohelp/delphi/delphirpc"
	"os"
	"path/filepath"
	r "reflect"
)

func main() {

	// delphiDir, golangDir, types, notifyTypes
	delphirpc.WriteSources("elco", delphirpc.SrcServices{
		Dir: filepath.Join(os.Getenv("DELPHIPATH"),
			"src", "github.com", "fpawel", "elco-gui", "api"),
		Types: []r.Type{
			r.TypeOf((*api.PartiesCatalogueSvc)(nil)),
			r.TypeOf((*api.LastPartySvc)(nil)),
			r.TypeOf((*api.ProductTypesSvc)(nil)),
			r.TypeOf((*api.PlaceFirmware)(nil)),
			r.TypeOf((*api.RunnerSvc)(nil)),
			r.TypeOf((*api.PdfSvc)(nil)),
			r.TypeOf((*api.PeerSvc)(nil)),
			r.TypeOf((*api.ConfigSvc)(nil)),
			r.TypeOf((*api.ProductsCatalogueSvc)(nil)),
		},
	}, delphirpc.SrcNotify{
		Dir: filepath.Join(os.Getenv("GOPATH"),
			"src", "github.com", "fpawel", "elco", "internal", "api", "notify"),
		Types: []delphirpc.NotifyServiceType{
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
		},
		PeerWindowClassName:   "TElcoMainForm",
		ServerWindowClassName: "ElcoServerWindow",
	})
}
