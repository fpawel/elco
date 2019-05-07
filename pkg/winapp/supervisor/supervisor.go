package supervisor

import (
	"bytes"
	"fmt"
	"github.com/fpawel/elco/pkg/panichook"
	"github.com/fpawel/elco/pkg/winapp"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"
)

// ExecuteProcess запускает exeFileName
// В каталоге исполняемого файла создаётся папка logs.
// Стандартный вывод запущенного процесса добавляется в конец файла logs\[2006-01-02].log
// Возвращаемое значение - сообщение о панике, либо пустая строка
func ExecuteProcess(exeFileName string, args ...string) string {

	log.SetFlags(log.Ltime)

	exeDir := filepath.Dir(exeFileName)
	t := time.Now()
	logDir := filepath.Join(exeDir, "logs")

	if err := winapp.EnsuredDirectory(logDir); err != nil {
		log.Fatal(err)
	}

	logFileName := filepath.Join(logDir, fmt.Sprintf("%s.log", t.Format("2006-01-02")))
	log.Println("log file:", logFileName)

	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		log.Println("close log file: ", logFileName, logFile.Close())
	}()

	cmd := exec.Command(exeFileName, args...)
	panicBuffer := bytes.NewBuffer(nil)
	cmd.Stderr = &redirectOutput{logFile: logFile, panicBuffer: panicBuffer}
	cmd.Stdout = &redirectOutput{logFile: logFile, panicBuffer: panicBuffer}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	err = cmd.Wait()
	if err == nil {
		return ""
	}

	log.SetOutput(redirectOutput2{logFile: logFile})
	log.Println(err)
	panicParsed := bytes.NewBuffer(nil)

	panicMsg := parsePanicMsg(panicBuffer.String())

	if err := panichook.DumpCrash(panicBuffer, panicParsed); err != nil {
		log.Println("panichook.DumpCrash:", err)
	}

	panicParsedStr := panicParsed.String()

	log.Println(panicParsedStr)
	return panicMsg + "\n\n" + panicParsedStr
}

type redirectOutput struct {
	logFile     *os.File
	panicBuffer *bytes.Buffer
	panic       bool
}

func (x *redirectOutput) Write(p []byte) (int, error) {

	defer ResetColor()

	_, _ = fmt.Fprint(x.logFile, time.Now().Format("15:04:05"), " ")
	nResult, err := x.logFile.Write(p)
	if err != nil {
		log.Fatal(err)
	}

	if !x.panic {
		Foreground(Green, true)
		fmt.Print(time.Now().Format("15:04:05"), " ")
	}

	if bytes.HasPrefix(p, []byte("panic:")) {
		x.panic = true
	}

	if x.panic {
		_, _ = x.panicBuffer.Write(p)
		Foreground(Red, true)
	} else {
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
	}

	_, _ = os.Stderr.Write(p)
	return nResult, nil
}

func parsePanicMsg(s string) string {
	xs := regexPanicMsg.FindStringSubmatch(s)
	if len(xs) > 1 {
		return xs[1]
	}
	return ""
}

var regexPanicMsg = regexp.MustCompile(`panic:\s*([^\n]+)\n`)

type redirectOutput2 struct {
	logFile *os.File
}

func (x redirectOutput2) Write(p []byte) (int, error) {
	_, _ = os.Stderr.Write(p)
	return x.logFile.Write(p)
}
