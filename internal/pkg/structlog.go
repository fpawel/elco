package pkg

import (
	"github.com/powerman/structlog"
)

func LogPrependSuffixKeys(log *structlog.Logger, args ...interface{}) *structlog.Logger {
	var keys []string
	for i, arg := range args {
		if i%2 == 0 {
			k, ok := arg.(string)
			if !ok {
				panic("key must be string")
			}
			keys = append(keys, k)
		}
	}
	return log.New(args...).PrependSuffixKeys(keys...)
}

func NewLogWithSuffixKeys(args ...interface{}) *structlog.Logger {
	return LogPrependSuffixKeys(structlog.New(), args...)
}
