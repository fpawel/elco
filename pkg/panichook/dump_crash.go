package panichook

import (
	"fmt"
	"github.com/maruel/panicparse/stack"
	"io"
)

func DumpCrash(in io.Reader, out io.Writer) error {
	// Optional: Check for GOTRACEBACK being set, in particular if there is only
	// one goroutine returned.
	c, err := stack.ParseDump(in, emptyWriter{}, true)
	if err != nil {
		return err
	}

	// Find out similar goroutine traces and group them into buckets.
	buckets := stack.Aggregate(c.Goroutines, stack.AnyValue)

	// Calculate alignment.
	srcLen := 0
	pkgLen := 0
	for _, bucket := range buckets {
		for _, line := range bucket.Signature.Stack.Calls {
			if l := len(line.SrcLine()); l > srcLen {
				srcLen = l
			}
			if l := len(line.Func.PkgName()); l > pkgLen {
				pkgLen = l
			}
		}
	}

	for _, bucket := range buckets {
		// Print the goroutine header.
		extra := ""
		if s := bucket.SleepString(); s != "" {
			extra += " [" + s + "]"
		}
		if bucket.Locked {
			extra += " [locked]"
		}
		if c := bucket.CreatedByString(false); c != "" {
			extra += " [Created by " + c + "]"
		}
		if _, err := fmt.Fprintf(out, "%d: %s%s\n", len(bucket.IDs), bucket.State, extra); err != nil {
			return err
		}

		// Print the stack lines.
		for _, line := range bucket.Stack.Calls {
			if _, err := fmt.Fprintf(out,
				"    %-*s %-*s %s(%s)\n",
				pkgLen, line.Func.PkgName(), srcLen, line.SrcLine(),
				line.Func.Name(), &line.Args); err != nil {
				return err
			}
		}
		if bucket.Stack.Elided {
			if _, err := fmt.Fprintf(out, "    (...)\n"); err != nil {
				return err
			}
		}
	}
	return nil
}

type emptyWriter struct {
}

func (_ emptyWriter) Write(b []byte) (int, error) {
	return len(b), nil
}
