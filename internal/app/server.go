package app

import (
	"github.com/fpawel/elco/internal/api"
	"github.com/powerman/must"
	"github.com/powerman/rpc-codec/jsonrpc2"
	"golang.org/x/sys/windows/registry"
	"io"
	"net"
	"net/http"
	"net/rpc"
)

func (x *App) startHttpServer() func() {

	// Server export an object of type ExampleSvc.
	for _, svcObj := range []interface{}{
		new(api.PartiesCatalogueSvc),
		new(api.LastPartySvc),
		new(api.ProductTypesSvc),
		api.NewProductFirmware(x),
		new(api.SettingsSvc),
		new(api.PdfSvc),
		&api.RunnerSvc{Runner: x},
	} {
		must.AbortIf(rpc.Register(svcObj))
	}

	// Server provide a HTTP transport on /rpc endpoint.
	http.Handle("/rpc", jsonrpc2.HTTPHandler(nil))

	srv := new(http.Server)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "hello world\n")
	})
	lnHTTP, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	addr := "http://" + lnHTTP.Addr().String()
	log.Info(addr)
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `elco\http`, registry.ALL_ACCESS)
	if err != nil {
		panic(err)
	}
	if err := key.SetStringValue("addr", addr); err != nil {
		panic(err)
	}
	log.ErrIfFail(key.Close)

	go func() {
		err := srv.Serve(lnHTTP)
		if err == http.ErrServerClosed {
			return
		}
		log.PrintErr(err)
		log.ErrIfFail(lnHTTP.Close)
	}()

	return func() {
		if err := srv.Shutdown(x.ctx); err != nil {
			log.PrintErr(err)
		}
	}
}
