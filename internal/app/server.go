package app

import (
	"fmt"
	"github.com/fpawel/elco/internal/pkg/must"
	"github.com/powerman/rpc-codec/jsonrpc2"
	"github.com/powerman/structlog"
	"net"
	"net/http"
	"os"
	"strconv"
)

func startHttpServer() func() {

	log := structlog.New()

	// Server provide a HTTP transport on /rpc endpoint.
	http.Handle("/rpc", jsonrpc2.HTTPHandler(nil))

	srv := new(http.Server)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		must.Write(w, []byte("hello world"))
	})

	port, errPort := strconv.Atoi(os.Getenv("ELCO_API_PORT"))
	if errPort != nil {
		log.Debug("finding free port to serve api")
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			panic(err)
		}
		port = ln.Addr().(*net.TCPAddr).Port
		must.PanicIf(os.Setenv("ELCO_API_PORT", strconv.Itoa(port)))
		must.PanicIf(ln.Close())
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	log.Debug("serve api: " + addr)

	lnHTTP, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	//addr := "http://" + lnHTTP.Addr().String()
	//log.Info(addr)
	//key, _, err := registry.CreateKey(registry.CURRENT_USER, `elco\http`, registry.ALL_ACCESS)
	//if err != nil {
	//	panic(err)
	//}
	//if err := key.SetStringValue("addr", addr); err != nil {
	//	panic(err)
	//}
	//log.ErrIfFail(key.Close)

	go func() {
		err := srv.Serve(lnHTTP)
		if err == http.ErrServerClosed {
			return
		}
		log.PrintErr(err)
		log.ErrIfFail(lnHTTP.Close)
	}()

	return func() {
		log.Debug("close http server")
		log.ErrIfFail(func() error {
			return srv.Shutdown(ctxApp)
		})
	}
}
