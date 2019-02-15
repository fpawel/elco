package comport

import (
	"context"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/tarm/serial"
	"time"
)

type Ports struct {
	ctx context.Context
	m   map[string]*Port
}

func NewPorts(ctx context.Context) *Ports {
	return &Ports{
		m:   map[string]*Port{},
		ctx: ctx,
	}
}

func (x *Ports) Close() (err error) {
	for k, p := range x.m {
		delete(x.m, k)
		if p.Opened() {
			if e := p.Close(); e != nil {
				err = multierror.Append(err, errors.Wrap(e, k))
			}
		}
	}
	return
}

func (x *Ports) Open(serialPortName string, baud int, bounceTimeout time.Duration) (*Port, error) {

	c := serial.Config{
		Name:        serialPortName,
		Baud:        baud,
		ReadTimeout: time.Millisecond,
	}

	if p, f := x.m[c.Name]; f {

		currC := p.Config()

		if currC == c {
			return p, nil
		}

		if p.Opened() {
			if err := p.Close(); err != nil {
				delete(x.m, c.Name)
				return nil, err
			}
		}
		if err := p.Open(serialPortName, baud, bounceTimeout, x.ctx); err != nil {
			delete(x.m, c.Name)
			return nil, err
		}

		return p, nil
	}
	p := new(Port)
	if err := p.Open(serialPortName, baud, bounceTimeout, x.ctx); err != nil {
		return nil, err
	}
	x.m[c.Name] = p
	return p, nil
}
