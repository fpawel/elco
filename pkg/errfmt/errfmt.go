package errfmt

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/sirupsen/logrus"
)

func Values(err error) (m logrus.Fields) {
	for k, v := range merry.Values(err) {
		if len(m) == 0 {
			m = logrus.Fields{}
		}
		m[fmt.Sprintf("%v", k)] = fmt.Sprintf("%v", v)
	}
	if len(m) == 0 {
		m = logrus.Fields{}
	}
	m["stack"] = merry.Stacktrace(err)
	return
}

func Format(err error, inclureStack bool) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	for k, v := range merry.Values(err) {
		k := fmt.Sprintf("%v", k)
		switch k {
		case "stack", "msg", "message", "time", "level", "work":
			continue
		default:
			s += "\n" + fmt.Sprintf("%s: %v", k, v)
		}
	}
	if inclureStack {
		s += "\n" + merry.Stacktrace(err)
	}
	return s
}

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
