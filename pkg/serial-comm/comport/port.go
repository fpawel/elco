package comport

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/pkg/errfmt"
	"github.com/fpawel/elco/pkg/serial-comm/comm"
	"github.com/fpawel/serial"
	"github.com/hako/durafmt"
	"strings"
	"sync"
	"time"
)

type Comm struct {
	Port   *Port
	Config comm.Config
}

func (x Comm) GetResponse(request []byte) ([]byte, error) {
	return x.Port.GetResponse(x.Config, request)
}

type Port struct {
	config serial.Config
	port   *serial.Port
	ctx    context.Context
	hook   Hook
	device string
}

type Entry struct {
	Request  []byte
	Response []byte
	Duration time.Duration
	Port     string
	Device   string
	Error    error
}

type Hook func(Entry)

func NewPort(device string, h Hook) *Port {
	return &Port{
		hook:   h,
		device: device,
	}
}

func (x *Port) SetHook(hook Hook) {
	x.hook = hook
}

func (x *Port) Config() serial.Config {
	return x.config
}

func (x *Port) Open(serialPortName string, baud int, bounceTimeout time.Duration, ctx context.Context) (err error) {
	if x.Opened() {
		return merry.New("already opened")
	}
	config := serial.Config{
		Name:        serialPortName,
		Baud:        baud,
		ReadTimeout: time.Millisecond,
	}

	if bounceTimeout == 0 {
		x.port, err = openPort(serial.Config{
			Name:        serialPortName,
			Baud:        baud,
			ReadTimeout: time.Millisecond,
		})
	} else {
		x.port, err = openPortWithBounceTimeout(config, bounceTimeout, ctx)
	}
	if err != nil {
		return
	}
	x.ctx = ctx
	x.config = config
	return
}

func openPortWithBounceTimeout(config serial.Config, bounceTimeout time.Duration, ctx context.Context) (port *serial.Port, err error) {
	ctx, _ = context.WithTimeout(ctx, bounceTimeout)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				if err == nil {
					err = ctx.Err()
				}
				return
			default:
				if port, err = openPort(config); err == nil {
					return
				}
			}
		}
	}()
	wg.Wait()
	return
}

func openPort(config serial.Config) (*serial.Port, error) {
	if err := CheckPortName(config.Name); err != nil {
		return nil, err
	}
	port, err := serial.OpenPort(&config)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "the system cannot find the file specified") {
			return nil, merry.New("нет СОМ порта с таким именем")
		}
		if strings.Contains(strings.ToLower(err.Error()), "access is denied") {
			return nil, merry.New("СОМ порт занят")
		}
		return nil, merry.Wrap(err)
	}
	return port, nil
}

func (x *Port) Opened() bool {
	return x.port != nil
}

func (x *Port) Write(buf []byte) (int, error) {
	if !x.Opened() {
		return 0, merry.Errorf("%s: was not opened", x.config.Name)
	}

	if err := x.port.Flush(); err != nil {
		return 0, merry.Wrap(err)
	}
	n, err := x.port.Write(buf)
	return n, merry.Wrap(err)
}

func (x *Port) Read(p []byte) (int, error) {
	if !x.Opened() {
		return 0, merry.Errorf("%s: was not opened", x.config.Name)
	}
	n, err := x.port.Read(p)
	return n, merry.Wrap(err)

}

func (x *Port) Close() error {
	if x.port == nil {
		return nil
	}
	err := x.port.Close()
	x.port = nil
	return err
}

func (x *Port) BytesToReadCount() (int, error) {
	var (
		errors   uint32
		commStat serial.CommStat
	)
	if err := x.port.ClearCommError(&errors, &commStat); err != nil {
		return 0, merry.WithMessage(err, "unable to get bytes to read count")
	}
	return int(commStat.InQue), nil
}

func (x *Port) GetResponse(commConfig comm.Config, request []byte) ([]byte, error) {
	//x.Port.GetResponse(request, x.Config)

	t := time.Now()
	response, err := comm.GetResponse(x.ctx, commConfig, x, x, request)
	duration := time.Since(t)

	if err == context.DeadlineExceeded {
		err = merry.WithMessage(context.DeadlineExceeded, "не отвечает")
	}

	if err != nil {
		err = errfmt.WithReqResp(err, request, response).
			WithValue("port", x.config.Name).
			WithValue("device", x.device).
			WithValue("duration", durafmt.Parse(duration))
	}

	if x.hook != nil {
		x.hook(Entry{
			Request:  request,
			Response: response,
			Duration: duration,
			Port:     x.config.Name,
			Error:    err,
			Device:   x.device,
		})
	}
	return response, err
}

func (x Entry) String() string {
	s := fmt.Sprintf("%s: %s: % X", x.Device, x.Port, x.Request)
	if len(x.Response) > 0 {
		s += fmt.Sprintf("-> % X", x.Response)
	}
	if x.Error != nil {
		s += ": " + x.Error.Error()
	}
	s += ": " + durafmt.Parse(x.Duration).String()
	return s
}
