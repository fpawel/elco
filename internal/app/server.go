package app

import (
	"github.com/fpawel/elco/internal/peer"
	"github.com/powerman/must"
	"github.com/powerman/rpc-codec/jsonrpc2"
	"github.com/powerman/structlog"
	"golang.org/x/sys/windows/registry"
	"net"
	"net/http"
)

func startHttpServer() func() {

	log := structlog.New()

	// Server provide a HTTP transport on /rpc endpoint.
	http.Handle("/rpc", jsonrpc2.HTTPHandler(nil))

	srv := new(http.Server)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		must.Write(w, []byte("hello world"))
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
		if err := srv.Shutdown(ctxApp); err != nil {
			log.PrintErr(err)
		}
	}
}

type peerNotifier struct{}

func (_ peerNotifier) OnStarted() {
	peer.InitPeer()
	cancelWorkFunc()
}

func (_ peerNotifier) OnClosed() {
	peer.ResetPeer()
	cancelWorkFunc()
}
