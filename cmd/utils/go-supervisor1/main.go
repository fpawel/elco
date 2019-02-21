package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"strings"
)

func main() {
	var args struct {
		ExeFileName string `arg:"positional" help:"file name to execute"`
		ExeArgs     string `arg:"-a" help:"command line arguments"`
	}
	arg.MustParse(&args)
	fmt.Println("ExeFileName:", args.ExeFileName)
	fmt.Println("ExeArgs:")
	for i, a := range strings.Split(args.ExeArgs, " ") {
		fmt.Printf("\t[%d] %q\n", i, a)
	}
}
