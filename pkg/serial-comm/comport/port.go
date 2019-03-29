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

type Port struct {
	config serial.Config
	port   *serial.Port

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

func NewPort(device string, config serial.Config, h Hook) *Port {
	if config.ReadTimeout == 0 {
		config.ReadTimeout = time.Millisecond
	}
	return &Port{
		hook:   h,
		config: config,
		device: device,
	}
}

func (x *Port) Config() serial.Config {
	return x.config
}

func (x *Port) Open(name string) error {
	if x.Opened() {
		return merry.New("already opened")
	}
	x.config.Name = name
	port, err := openPort(x.config)
	if err == nil {
		x.port = port
	}
	if err != nil {
		err = merry.Append(err, name)
	}
	return err
}

func (x *Port) OpenWithDebounce(name string, bounceTimeout time.Duration, ctx context.Context) error {
	if x.Opened() {
		return merry.New("already opened")
	}
	x.config.Name = name
	port, err := openPortWithBounceTimeout(x.config, bounceTimeout, ctx)
	if err == nil {
		x.port = port
	}
	return err
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

func (x *Port) GetResponse(request []byte, commConfig comm.Config, ctx context.Context, prs comm.ResponseParser) ([]byte, error) {
	//x.Port.GetResponse(request, x.Config)

	t := time.Now()
	response, err := comm.GetResponse(comm.Request{
		Bytes:              request,
		Config:             commConfig,
		ReadWriter:         x,
		BytesToReadCounter: x,
		ResponseParser:     prs,
	}, ctx)
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
		go x.hook(Entry{
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
		s += fmt.Sprintf(" -> % X", x.Response)
	}
	if x.Error != nil {
		s += ": " + x.Error.Error()
	}
	s += ": " + durafmt.Parse(x.Duration).String()
	return s
}
