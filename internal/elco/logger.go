package elco

import "github.com/sirupsen/logrus"

var Logger = &logrus.Logger{
	Formatter: &logrus.TextFormatter{TimestampFormat: "15:04:05.000"},
	Level:     logrus.InfoLevel,
	Out:       ColorStdOutWriter{},
}
