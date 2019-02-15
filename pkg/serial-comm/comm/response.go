package comm

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/hako/durafmt"
	"io"
	"time"
)

func GetResponse(ctx context.Context, config Config, readWriter io.ReadWriter, request []byte) ([]byte, error) {
	if config.MaxAttemptsRead > 10 {
		config.MaxAttemptsRead = 10
	}

	return responseReader{
		config:     config,
		readWriter: readWriter,
		request:    request,
	}.getResponse(ctx, 0)
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

func DefaultConfig() Config {
	return Config{
		ReadTimeoutMillis:     1000,
		ReadByteTimeoutMillis: 50,
	}
}

type responseReader struct {
	request    []byte
	readWriter io.ReadWriter
	config     Config
}

func (x responseReader) getResponse(parentCtx context.Context, attempt int) ([]byte, error) {

	if err := x.write(); err != nil {
		return nil, err
	}

	ctx, _ := context.WithTimeout(parentCtx, x.config.ReadTimeout())
	chResponse := make(chan []byte)
	chError := make(chan error)

	go x.doGetResponse(ctx, chResponse, chError)

	select {

	case response := <-chResponse:
		return response, nil

	case err := <-chError:
		return nil, err

	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded && attempt < x.config.MaxAttemptsRead {
			ctx, _ = context.WithTimeout(parentCtx, x.config.ReadTimeout())
			chNextAttempt := time.After(x.config.ReadByteTimeout())
			for {
				select {
				case <-parentCtx.Done():
					return nil, parentCtx.Err()
				case <-chNextAttempt:
					return x.getResponse(parentCtx, attempt+1)
				}
			}
		} else {
			return nil, ctx.Err()
		}
	}
}

func (x responseReader) write() error {

	t := time.Now()
	writtenCount, err := x.readWriter.Write(x.request)
	for ; err == nil && writtenCount == 0 && time.Since(t) < x.config.ReadTimeout(); writtenCount, err = x.readWriter.Write(x.request) {
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

func (x responseReader) doGetResponse(ctx context.Context, chResponse chan []byte,
	chError chan error) {

	timerReady := time.NewTimer(0)
	<-timerReady.C
	var response []byte
	for {
		select {

		case <-ctx.Done():
			return

		case <-timerReady.C:
			chResponse <- response
			return

		default:
			// пытаться считать байт ответа
			b := []byte{0}
			readCount, err := x.readWriter.Read(b)
			if err != nil {
				chError <- err
				return
			}
			if readCount == 0 {
				continue
			}
			response = append(response, b[0])
			timerReady.Stop()
			timerReady = time.NewTimer(x.config.ReadByteTimeout())
		}
	}
}
