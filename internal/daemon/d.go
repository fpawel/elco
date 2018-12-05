package daemon

import (
	"context"
	"fmt"
	"github.com/Microsoft/go-winio"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/app/config"
	"github.com/fpawel/elco/internal/crud"
	"github.com/fpawel/goutils/copydata"
	"github.com/fpawel/goutils/serial/comport"
	"github.com/hashicorp/go-multierror"
	"github.com/lxn/win"
	"github.com/pkg/errors"
	"github.com/powerman/rpc-codec/jsonrpc2"
	"net"
	"net/rpc"
	"sync"
	"sync/atomic"
)

type D struct {
	c    crud.DBContext         // база данных sqlite
	w    *copydata.NotifyWindow // окно для отправки сообщений WM_COPYDATA дельфи-приложению
	sets *config.Sets

	comports struct {
		comport.Ports
		sync.WaitGroup
		context.Context
		cancel func()
	}
}

const (
	PipeName              = `\\.\pipe\elco`
	ServerWindowClassName = "ElcoServerWindow"
	PeerWindowClassName   = "TElcoMainForm"
)

func New() *D {
	c := crud.NewDBContext(nil)
	sets, err := config.OpenSets(c.LastParty())

	if err != nil {
		fmt.Println("sets:", err)
	}

	x := &D{
		c:    c,
		sets: sets,
		w:    copydata.NewNotifyWindow(ServerWindowClassName, PeerWindowClassName),
	}
	x.comports.Ports = make(comport.Ports)
	x.registerRPCServices()
	return x
}

func (x *D) Run(closeOnDisconnect bool) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(2)
	ln := mustPipeListener()
	// цикл RPC сервера
	go func() {
		defer wg.Done()
		defer x.w.CloseWindow()
		x.serveRPC(ln, ctx, closeOnDisconnect)
	}()
	// цикл оконных сообщений
	runWindowMessageLoop()

	x.comports.cancel()
	x.comports.WaitGroup.Wait()

	cancel()

	if err := ln.Close(); err != nil {
		fmt.Println("close pipe listener error:", err)
	}
	wg.Wait()
}

func (x *D) Close() (result error) {

	if err := x.c.Close(); err != nil {
		result = multierror.Append(result, errors.Wrap(err, "close sqlite data base"))
	}

	if err := x.sets.Save(); err != nil {
		result = multierror.Append(result, errors.Wrap(err, "save config"))
	}

	return
}

func (x *D) serveRPC(ln net.Listener, ctx context.Context, closeOnDisconnectPeer bool) {
	count := int32(0)
	for {
		switch conn, err := ln.Accept(); err {
		case nil:
			go func() {
				atomic.AddInt32(&count, 1)
				jsonrpc2.ServeConnContext(ctx, conn)
				if atomic.AddInt32(&count, -1) == 0 && closeOnDisconnectPeer {
					return
				}
			}()
		case winio.ErrPipeListenerClosed:
			return
		default:
			fmt.Println("rpc pipe error:", err)
			return
		}
	}
}

func (x *D) registerRPCServices() {
	for _, svcObj := range []interface{}{
		api.NewPartiesCatalogue(x.c.PartiesCatalogue()),
		api.NewLastParty(x.c.LastParty()),
		api.NewProductTypes(x.c.ProductTypes()),
		api.NewProductFirmware(x.c.ProductFirmware()),
		api.NewSetsSvc(x.sets),
	} {
		if err := rpc.Register(svcObj); err != nil {
			panic(err)
		}
	}
}

func mustPipeListener() net.Listener {
	ln, err := winio.ListenPipe(PipeName, nil)
	if err != nil {
		panic(err)
	}
	return ln
}

func runWindowMessageLoop() {
	for {
		var msg win.MSG
		if win.GetMessage(&msg, 0, 0, 0) == 0 {
			break
		}
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}
}
