package app

import (
	"bytes"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/gohelp/intrng"
	"path/filepath"
	"runtime"
	"sort"
)

func formatProducts(products []data.Product) (s string) {
	var places []int
	for _, p := range products {
		places = append(places, p.Place)
	}
	sort.Ints(places)
	intrng.IterateOverRanges(places, func(n int, k int) {
		if s != "" {
			s += " "
		}
		if n == k {
			s += data.FormatPlace(n)
		} else {
			s += fmt.Sprintf("%s-%s", data.FormatPlace(n), data.FormatPlace(k))
		}
	})
	return s
}

// stacktrace returns the error's stacktrace as a string formatted
// the same way as golangs runtime package.
// If e has no stacktrace, returns an empty string.
func merryStacktrace(e error) string {

	s := merry.Stack(e)
	if len(s) > 0 {
		buf := bytes.Buffer{}
		for i, fp := range s {
			fnc := runtime.FuncForPC(fp)
			if fnc != nil {
				f, l := fnc.FileLine(fp)
				name := filepath.Base(fnc.Name())
				ident := " "
				if i > 0 {
					ident = "\t"
				}

				buf.WriteString(fmt.Sprintf("%s%s:%d %s\n", ident, f, l, name))
			}
		}
		return buf.String()
	}
	return ""
}
