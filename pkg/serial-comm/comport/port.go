package comport

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/pkg/serial-comm/comm"
	"github.com/hako/durafmt"
	"github.com/sirupsen/logrus"
	"github.com/tarm/serial"
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
	config            serial.Config
	port              *serial.Port
	ctx               context.Context
	hook              Hook
	request, response []byte
	duration          time.Duration
	err               error
	muLogger          sync.Mutex
	logger            *logrus.Logger
	logFields         logrus.Fields
}

type LastWork struct {
	Request, Response []byte
	Duration          time.Duration
	Port              string
	Error             error
}

type Hook func(LastWork)

func NewPort(logFields logrus.Fields, h Hook) *Port {
	return &Port{
		hook:      h,
		logFields: logFields,
	}
}

func (x *Port) LastWork() LastWork {
	return LastWork{
		Error:    x.err,
		Port:     x.config.Name,
		Duration: x.duration,
		Request:  x.request,
		Response: x.response,
	}
}

func (x *Port) SetHook(hook Hook) {
	x.hook = hook
}

func (x *Port) Config() serial.Config {
	return x.config
}

func (x *Port) SetLogger(logger *logrus.Logger) {
	x.muLogger.Lock()
	x.logger = logger
	x.muLogger.Unlock()
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

func (x *Port) GetResponse(commConfig comm.Config, request []byte) ([]byte, error) {
	//x.Port.GetResponse(request, x.Config)

	x.request = request
	t := time.Now()
	x.response, x.err = comm.GetResponse(x.ctx, commConfig, x, request)
	x.duration = time.Since(t)
	if x.err != nil {
		x.err = x.WrapError(x.err)
	}
	if x.hook != nil {
		x.hook(x.LastWork())
	}

	x.muLogger.Lock()
	logger := x.logger
	x.muLogger.Unlock()

	if x.err != nil && logger == nil {
		logger = logrus.StandardLogger()
	}
	if logger != nil {
		entry := logrus.NewEntry(logger)
		if x.err == nil {
			entry.Data = x.logKeyValues()
			entry.Infoln()
		} else {
			entry.Errorln(merry.Details(x.err))
		}
	}
	return x.response, x.err
}

func (x *Port) NewError(msg string) error {
	return x.WrapError(merry.New(msg))
}

func (x *Port) Errorf(fmt string, a ...interface{}) error {
	return x.WrapError(merry.Errorf(fmt, a...))
}

func (x *Port) WrapError(e error) error {

	var err merry.Error
	switch e {
	case nil:
		return nil
	case context.Canceled:
		return context.Canceled
	case context.DeadlineExceeded:
		err = merry.WithMessage(context.DeadlineExceeded, "не отвечает")
	default:
		err = merry.Wrap(e)
	}
	for k, v := range x.logKeyValues() {
		err = err.WithValue(k, v)
	}
	return err
}

func (x *Port) logKeyValues() logrus.Fields {
	m := logrus.Fields{
		"port":     x.config.Name,
		"request":  fmt.Sprintf("% X", x.request),
		"duration": durafmt.Parse(x.duration),
	}
	if len(x.response) > 0 {
		m["response"] = fmt.Sprintf("% X", x.response)
	}
	for k, v := range x.logFields {
		m[k] = v
	}
	return m
}

func (x LastWork) String() string {
	var strErr, strResponse string
	if x.Error != nil {
		strErr = " : " + x.Error.Error()
	}
	if len(x.Response) > 0 {
		strResponse = fmt.Sprintf(": % X ", x.Response)
	}
	return fmt.Sprintf("% X %s : %s%s", x.Request, strResponse, durafmt.Parse(x.Duration), strErr)
}
