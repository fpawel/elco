package ccolor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"
)

type Output struct {
	f  *os.File
	ln bool
}

func NewWriter(f *os.File) io.Writer {
	return &Output{f: f, ln: true}
}

func (x *Output) Write(p []byte) (int, error) {
	if x.ln {
		Foreground(Green, true)
		_, _ = fmt.Fprint(x.f, time.Now().Format("15:04:05"), " ")
		fields := bytes.Fields(p)
		if len(fields) > 1 {
			switch string(fields[1]) {
			case "ERR":
				Foreground(Red, true)
			case "WRN":
				Foreground(Yellow, true)
			case "inf":
				Foreground(White, true)
			default:
				Foreground(White, false)
			}
		}
		defer ResetColor()
	}
	x.ln = bytes.HasSuffix(p, []byte("\n"))
	return x.f.Write(p)
}
