package logfile

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/fpawel/elco/internal/pkg"
	"github.com/fpawel/elco/internal/pkg/must"
	"github.com/jmoiron/sqlx"
	"github.com/powerman/structlog"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

func NewOutput() io.WriteCloser {
	must.EnsureDir(logDir)
	f, err := OpenCurrentFile()
	must.AbortIf(err)
	return &output{f: f}
}

func OpenCurrentFile() (*os.File, error) {
	return os.OpenFile(currentFilename(), os.O_CREATE|os.O_APPEND, 0666)
}

func ListDays() []time.Time {
	r := regexp.MustCompile(`\d\d\d\d-\d\d-\d\d`)
	m := make(map[time.Time]struct{})
	_ = filepath.Walk(logDir, func(path string, f os.FileInfo, _ error) error {
		if f == nil || f.IsDir() {
			return nil
		}
		xs := r.FindStringSubmatch(f.Name())
		if len(xs) == 0 {
			return nil
		}
		t, err := time.ParseInLocation("2006-01-02", xs[0], time.Local)
		if err != nil {
			return nil
		}
		m[daytime(t)] = struct{}{}
		return nil
	})
	var days []time.Time
	for t := range m {
		days = append(days, t)
	}
	sort.Slice(days, func(i, j int) bool {
		return days[i].Before(days[j])
	})
	return days
}

type Entry struct {
	Time time.Time `db:"tm"`
	Line string    `db:"tx"`
}

func Read(t time.Time, filter string) ([]Entry, error) {
	t = daytime(t)

	r := regexp.MustCompile(`\d\d\d\d-\d\d-\d\d`)

	db := must.OpenSqliteDBx(":memory:")
	defer log.ErrIfFail(db.Close)

	db.MustExec(`
CREATE TABLE entry(
	tm DATETIME  NOT NULL ,
	tx TEXT NOT NULL    
)`)

	filter = strings.TrimSpace(filter)
	if len(filter) != 0 {
		filter = " WHERE " + filter
	}

	var entries []Entry
	if err := db.Select(&entries, "SELECT * FROM entry"+filter); err != nil {
		return nil, err
	}

	_ = filepath.Walk(logDir, func(path string, f os.FileInfo, _ error) error {
		if f == nil || f.IsDir() {
			return nil
		}
		xs := r.FindStringSubmatch(f.Name())
		if len(xs) == 0 {
			return nil
		}
		fileTime, err := time.ParseInLocation("2006-01-02", xs[0], time.Local)
		if err != nil {
			return nil
		}
		if fileTime == t {
			readEntries(path, t, db)
		}
		return nil
	})

	if err := db.Select(&entries, "SELECT * FROM entry"+filter+" ORDER BY tm"); err != nil {
		return nil, err
	}
	return entries, nil
}

type output struct {
	f *os.File
	b bytes.Buffer
}

func (x *output) Close() error {
	return x.f.Close()
}

func (x *output) Write(p []byte) (int, error) {
	if !bytes.HasSuffix(p, []byte("\n")) {
		x.b.Write(p)
	} else {
		if _, err := fmt.Fprint(x.f, time.Now().Format("15:04:05.000"), " "); err != nil {
			log.PrintErr(err)
		}
		if _, err := x.b.WriteTo(x.f); err != nil {
			log.PrintErr(err)
		}
		if _, err := x.f.Write(p); err != nil {
			log.PrintErr(err)
		}
	}
	return len(p), nil
}

func currentFilename() string {
	return filename(daytime(time.Now()), "")
}

func readEntries(filename string, dayTime time.Time, db *sqlx.DB) {
	file, err := os.Open(filename)

	if os.IsNotExist(err) {
		return
	}

	if err != nil {
		log.PrintErr(err, "file", filepath.Base(filename))
		return
	}
	defer log.ErrIfFail(file.Close)

	var (
		lineNumber int
		scanner    = bufio.NewScanner(file)
	)

	for scanner.Scan() {
		line := scanner.Text()
		var ent Entry
		if err := parseEntry(dayTime, line, &ent); err != nil {
			log.PrintErr(err,
				"line", fmt.Sprintf("%d:`%s`", lineNumber, line),
				"file", filepath.Base(filename))
			continue
		}
		db.MustExec(`INSERT INTO entry(tm, tx) VALUES (?,?)`, ent.Time, ent.Line)

		lineNumber++
	}
	must.AbortIf(scanner.Err())
}

func parseEntry(dayTime time.Time, line string, ent *Entry) error {
	if len(line) == 0 {
		return errors.New("empty line")
	}
	for i := range line {
		if line[i] == ' ' {
			if i+1 == len(line) {
				return errors.New("wrong format")
			}
			if t, err := time.Parse("2006-01-02 15:04:05.000", dayTime.Format("2006-01-02")+" "+line[:i]); err != nil {
				return err
			} else {
				ent.Line = line[i+1:]
				ent.Time = t
				return nil
			}
		}
	}
	return errors.New("wrong format")
}

func filename(t time.Time, suffix string) string {
	if err := pkg.EnsureDir(logDir); err != nil {
		panic(err)
	}
	return filepath.Join(logDir, fmt.Sprintf("%s%s.log", t.Format("2006-01-02"), suffix))
}

func daytime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

var (
	log    = structlog.New()
	logDir = filepath.Join(filepath.Dir(os.Args[0]), "logs")
)
