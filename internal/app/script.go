package app

import (
	"bufio"
	"fmt"
	"github.com/ansel1/merry"
	parsec "github.com/prataprc/goparsec"
	"strconv"
	"strings"
	"time"
)

func (_ runner) RunScript(script string) error {
	var works []workFunc
	if err := parseScript(script, &works); err != nil {
		return err
	}
	runWork("сценарий", func(x worker) error {
		for _, work := range works {
			if x.ctx.Err() != nil {
				return x.ctx.Err()
			}
			if err := work(x); err != nil {
				return err
			}
		}
		return nil
	})
	return nil
}

type workFunc = func(worker) error

func parseScript(script string, works *[]workFunc) error {

	add := func(f workFunc) {
		*works = append(*works, f)
	}

	ast := parsec.NewAST("script", 100)
	pFloat := parsec.Token(`[+-]?([0-9]+\.[0-9]*|\.[0-9]+|[0-9]+)`, "FLOAT")
	pEnd := ast.End("END")
	pInstr := ast.OrdChoice(
		"INSTR", nil,
		ast.And("TEMP",
			func(name string, s parsec.Scanner, node parsec.Queryable) parsec.Queryable {
				t, _ := strconv.ParseFloat(node.GetChildren()[1].GetValue(), 64)
				add(func(w worker) error {
					return setupTemperature(w, t)
				})
				log.Println("temperature", t)
				return node
			},
			parsec.Atom("temperature", "TEMP-KWD"),
			pFloat,
			pEnd,
		),
		ast.And("GAS",
			func(name string, s parsec.Scanner, node parsec.Queryable) parsec.Queryable {
				gas, _ := strconv.ParseInt(node.GetChildren()[1].GetValue(), 10, 8)
				add(func(w worker) error {
					return w.switchGas(int(gas))
				})
				log.Println("gas", gas)
				return node
			},
			parsec.Atom("gas", "GAS-KWD"),
			parsec.Token(`[1-3]+`, "GAS-VALVE"),
			pEnd,
		),
		ast.And("PAUSE",
			func(name string, s parsec.Scanner, node parsec.Queryable) parsec.Queryable {
				n, _ := strconv.ParseInt(node.GetChildren()[1].GetValue(), 10, 64)
				add(func(w worker) error {
					return delay(w, time.Minute*time.Duration(n), fmt.Sprintf("пауза %d минут", n))
				})
				log.Println("pause", n)
				return node
			},
			parsec.Atom("pause", "PAUSE-KWD"),
			parsec.Token(`[0-9]+`, "PAUSE-VALUE"),
			pEnd,
		),
		ast.And("SAVE",
			func(name string, s parsec.Scanner, node parsec.Queryable) parsec.Queryable {
				add(readSaveForLastTemperatureGas)
				log.Println("save")
				return node
			},
			parsec.Atom("save", "SAVE-KWD"),
			pEnd,
		),
	)

	lineScanner := bufio.NewScanner(strings.NewReader(script))
	lineNo := 1
	for lineScanner.Scan() {
		str := strings.TrimSpace(lineScanner.Text())
		ast.Reset()
		scanner := parsec.NewScanner([]byte(str))
		node, _ := ast.Parsewith(pInstr, scanner)
		if node == nil {
			log.PrintErr("parse error: "+str, "line", lineNo)
			ast.Reset()
			ast.SetDebug()
			ast.Parsewith(pInstr, parsec.NewScanner([]byte(str)))
			return merry.Errorf("ошибка в строке %d: %q\n", lineNo, str)
		}
		lineNo++
	}
	return nil
}
