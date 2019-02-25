package comm

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/hako/durafmt"
	"io"
	"time"
)

type BytesToReadCounter interface {
	io.ReadWriter
	BytesToReadCount() (int, error)
}

func GetResponse(ctx context.Context, config Config, readWriter io.ReadWriter,
	bytesToReadCounter BytesToReadCounter, request []byte) ([]byte, error) {
	return reader{
		bc:      bytesToReadCounter,
		rw:      readWriter,
		config:  config,
		request: request,
	}.getResponse(ctx)
}

type Config struct {
	ReadTimeoutMillis     int `toml:"read_timeout" comment:"таймаут получения ответа, мс"`
	ReadByteTimeoutMillis int `toml:"read_byte_timeout" comment:"таймаут окончания ответа, мс"`
	MaxAttemptsRead       int `toml:"max_attempts_read" comment:"число попыток получения ответа"`
}

func (x Config) ReadTimeout() time.Duration {
	return time.Duration(x.ReadTimeoutMillis) * time.Millisecond
}

func (x Config) ReadByteTimeout() time.Duration {
	return time.Duration(x.ReadByteTimeoutMillis) * time.Millisecond
}

type reader struct {
	bc      BytesToReadCounter
	rw      io.ReadWriter
	request []byte
	config  Config
}

type result struct {
	response []byte
	err      error
}

func (x reader) getResponse(parentCtx context.Context) ([]byte, error) {

	for attempt := 0; attempt < x.config.MaxAttemptsRead; attempt++ {
		if err := x.write(); err != nil {
			return nil, err
		}
		ctx, _ := context.WithTimeout(parentCtx, x.config.ReadTimeout())
		c := make(chan result)

		go x.waitForResponse(ctx, c)

		select {

		case r := <-c:

			if r.err == nil {
				return r.response, nil
			}

			return nil, merry.WithValue(r.err, "attempt", attempt)

		case <-ctx.Done():

			switch ctx.Err() {

			case context.DeadlineExceeded:
				continue

			case context.Canceled:
				return nil, merry.WithMessage(context.Canceled, "прервано")

			default:
				return nil, merry.WithValue(ctx.Err(), "attempt", attempt)
			}
		}
	}

	return nil, merry.WithMessage(context.DeadlineExceeded, "не отвечает").
		WithValue("attempt", x.config.MaxAttemptsRead).
		WithValue("timeout", durafmt.Parse(x.config.ReadTimeout())).
		WithValue("read_byte_timeout", durafmt.Parse(x.config.ReadByteTimeout()))

}

func (x reader) write() error {

	t := time.Now()
	writtenCount, err := x.rw.Write(x.request)
	for ; err == nil && writtenCount == 0 && time.Since(t) < x.config.ReadTimeout(); writtenCount, err = x.rw.Write(x.request) {
		// COMPORT PENDING
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil {
		return err
	}

	if writtenCount != len(x.request) {

		return merry.Errorf("записано %d из %d байт, %s", writtenCount, len(x.request), durafmt.Parse(time.Since(t)))
	}
	return err
}

func (x reader) waitForResponse(ctx context.Context, c chan result) {

	var response []byte
	ctxReady := context.Background()

	for {
		select {

		case <-ctx.Done():
			return

		case <-ctxReady.Done():
			c <- result{response, nil}
			return

		default:
			bytesToReadCount, err := x.bc.BytesToReadCount()
			if err != nil {
				c <- result{response, merry.Wrap(err)}
				return
			}

			if bytesToReadCount == 0 {
				time.Sleep(time.Millisecond)
				continue
			}
			b, err := x.read(bytesToReadCount)
			if err != nil {
				c <- result{response, merry.WithMessagef(err, "[% X]", response)}
				return
			}
			response = append(response, b...)
			ctx = context.Background()
			ctxReady, _ = context.WithTimeout(context.Background(), x.config.ReadByteTimeout())
		}
	}
}

func (x reader) read(bytesToReadCount int) ([]byte, error) {
	b := make([]byte, bytesToReadCount)
	readCount, err := x.rw.Read(b)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	if readCount != bytesToReadCount {
		return nil, merry.Errorf("await %d bytes, %d read", bytesToReadCount, readCount)
	}
	return b, nil
}
