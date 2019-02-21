package main

import (
	"bytes"
	"fmt"
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/elco/pkg/winapp"
	"github.com/fpawel/goutils/panichook"
	"github.com/go-logfmt/logfmt"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func main() {

	elco.Logger.SetLevel(logrus.InfoLevel)
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(elco.Logger.Formatter)
	logrus.SetOutput(elco.Logger.Out)
	logrus.SetReportCaller(true)

	exeFileName, err := elco.CurrentDirOrProfileFileName("elco.exe")
	if err != nil {
		logrus.Fatal(err)
	}
	exeDir := filepath.Dir(exeFileName)
	t := time.Now()
	logDir := filepath.Join(exeDir, "logs",
		fmt.Sprintf("%d", t.Year()),
		fmt.Sprintf("%2d", t.Month()))
	if err := winapp.EnsuredDirectory(logDir); err != nil {
		logrus.Fatal(err)
	}
	logFileName := filepath.Join(logDir, fmt.Sprintf("%2d.log", t.Day()))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND, 0666)
	if err := winapp.EnsuredDirectory(logDir); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			logrus.Fatal(err)
		}
	}()

	var output bytes.Buffer
	cmd := exec.Command(exeFileName)
	cmd.Stderr = &redirectOutput{buff: &output}
	cmd.Stdout = &redirectOutput{buff: &output}

	if err := cmd.Start(); err != nil {
		logrus.Panic(err)
	}

	if err := cmd.Wait(); err != nil {
		panicBuffer := bytes.NewBuffer(nil)
		if err := panichook.DumpCrash(&output, panicBuffer); err != nil {
			logrus.Panic(err)
		}
		if _, err = cmd.Stderr.Write(panicBuffer.Bytes()); err != nil {
			logrus.Panic(err)
		}
	}
}

type redirectOutput struct {
	buff          *bytes.Buffer
	logFile       *os.File
	panicOccurred bool
}

func (x *redirectOutput) Write(p []byte) (int, error) {

	br := bytes.NewReader(p)
	d := logfmt.NewDecoder(br)
	entry := logrus.NewEntry(elco.Logger)
	entry.Level = logrus.InfoLevel
	entry.Data = logrus.Fields{}

	for d.ScanRecord() {
		for d.ScanKeyval() {
			k, v := string(d.Key()), string(d.Value())
			if k == "level" {
				var err error
				entry.Level, err = logrus.ParseLevel(v)
				if err != nil {
					logrus.Println(err, ":", string(p))
				}
			} else {
				entry.Data[k] = v
			}
		}
	}
	if bytes.HasPrefix(p, []byte("panic:")) {
		x.panicOccurred = true
		entry.Level = logrus.ErrorLevel
	}
	x.buff.Write(p)
	entry.Logln(entry.Level)
	return x.logFile.Write(p)
}
