package elco

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-logfmt/logfmt"
	"github.com/sirupsen/logrus"
	"strings"
)

var Logger = &logrus.Logger{
	Formatter: &logrus.TextFormatter{TimestampFormat: "15:04:05.000"},
	Level:     logrus.InfoLevel,
	Out:       colorStdOutWriter{},
}

type colorStdOutWriter struct{}

func (x colorStdOutWriter) Write(p []byte) (int, error) {
	d := logfmt.NewDecoder(bytes.NewReader(p))
	s := strings.TrimSpace(string(p))
	for d.ScanRecord() {
		for d.ScanKeyval() {
			if string(d.Key()) == "level" {
				value := string(d.Value())
				switch value {
				case "error", "panic", "fatal":
					return color.New(color.FgRed, color.Bold).Println(s)
				case "warn", "warning":
					return color.New(color.FgHiMagenta, color.Bold).Println(s)
				case "info":
					return color.New(color.FgHiCyan).Println(s)
				}
			}
		}
	}
	return fmt.Println(s)
}
