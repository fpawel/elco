package errfmt

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/sirupsen/logrus"
)

func WithReqResp(err error, request, response []byte) merry.Error {
	if err == nil {
		return nil
	}
	merryErr := merry.Wrap(err).
		WithValue("request", fmt.Sprintf("% X", request))
	if len(response) > 0 {
		merryErr = merryErr.WithValue("response", fmt.Sprintf("% X", response))
	}
	return merryErr
}

func WithReqRespMsg(request, response []byte, msg string) merry.Error {
	return WithReqResp(merry.New(msg), request, response)
}

func WithReqRespMsgf(request, response []byte, fmt string, args ...interface{}) merry.Error {
	return WithReqResp(merry.Errorf(fmt, args...), request, response)
}

func LogDetails(err error, entry *logrus.Entry) {
	if err == nil {
		return
	}
	msg := merry.Message(err)

	for key, value := range merry.Values(err) {
		if key == "cause" || key == "message" || key == "stack" {
			continue
		}
		if len(entry.Data) == 0 {
			entry.Data = make(logrus.Fields)
		}
		entry.Data[fmt.Sprintf("%v", key)] = value
	}
	entry.Errorln(msg)

	s := merry.Stacktrace(err)
	if s != "" {
		logrus.Errorln(s)
	}

	if c := merry.Cause(err); c != nil {
		logrus.Errorln("Caused By:")
		LogDetails(c, entry)
	}
}
