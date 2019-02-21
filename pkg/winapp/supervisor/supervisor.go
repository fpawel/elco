package supervisor

import (
	"bytes"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fatih/color"
	"github.com/fpawel/elco/pkg/panichook"
	"github.com/fpawel/elco/pkg/winapp"
	"github.com/go-logfmt/logfmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func ExecuteProcess(exeFileName string, args ...string) (string, error) {
	exeFileName, err := winapp.CurrentDirOrProfileFileName(".elco", "elco.exe")
	if err != nil {
		return "", merry.Wrap(err)
	}
	exeDir := filepath.Dir(exeFileName)
	t := time.Now()
	logDir := filepath.Join(exeDir, "logs")
	if err := winapp.EnsuredDirectory(logDir); err != nil {
		return "", merry.Wrap(err)
	}
	logFileName := filepath.Join(logDir, fmt.Sprintf("%s.log", t.Format("2006-01-02")))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND, 0666)
	if err := winapp.EnsuredDirectory(logDir); err != nil {
		return "", merry.WithMessage(err, logFileName)
	}
	defer func() {
		log.Println("close log file:", logFileName, logFile.Close())
	}()

	cmd := exec.Command(exeFileName, args...)
	panicBuffer := bytes.NewBuffer(nil)
	cmd.Stderr = &redirectOutput{logFile: logFile, panicBuffer: panicBuffer}
	cmd.Stdout = &redirectOutput{logFile: logFile, panicBuffer: panicBuffer}

	if err := cmd.Start(); err != nil {
		return "", merry.Wrap(err)
	}

	if err := cmd.Wait(); err != nil {
		panicStr := panicBuffer.String()
		panicParsed := bytes.NewBuffer(nil)
		if err := panichook.DumpCrash(panicBuffer, panicParsed); err != nil {
			return "", merry.Wrap(err)
		}
		parsedStr := panicParsed.String()
		log.Println(panicStr)
		log.Println(parsedStr)

		if _, err := fmt.Fprintf(logFile, "time=%s level=panic msg=%q source=%q",
			time.Now().Format("15:04:05.000"), parsedStr, panicStr); err != nil {
			return "", merry.Wrap(err)
		}
		return parsedStr + "\n\n" + panicStr, nil
	}
	return "", nil
}

type redirectOutput struct {
	logFile     *os.File
	panicBuffer *bytes.Buffer
	panic       bool
}

func (x *redirectOutput) Write(p []byte) (int, error) {
	if bytes.HasPrefix(p, []byte("panic:")) {
		x.panic = true
	}
	line := string(p)
	if x.panic {
		return x.panicBuffer.Write(p)
	}

	br := bytes.NewReader(p)
	d := logfmt.NewDecoder(br)
	c := color.New(color.FgHiWhite, color.Bold)
	for d.ScanRecord() {
		for d.ScanKeyval() {
			k, v := string(d.Key()), string(d.Value())
			if k == "level" {
				switch v {
				case "fatal", "error", "panic":
					c = color.New(color.FgRed, color.Bold)
				case "warn", "warning":
					c = color.New(color.FgHiMagenta, color.Bold)
				case "info":
					color.New(color.FgHiCyan)
				}
			}
		}
	}
	if n, err := x.logFile.Write(p); err != nil {
		return n, err
	}
	return c.Print(line)
}
