package app

import (
	"bufio"
	"fmt"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	parsec "github.com/prataprc/goparsec"
	"strconv"
	"strings"
	"time"
)

func (_ runner) RunScript(script string) (r api.RunScriptResult) {
	var instructions []instruction
	r.ErrorLine, r.ErrorLineno = parseScript(script, &instructions)
	if r.ErrorLineno >= 0 {
		return
	}
	runWork("сценарий", func(x worker) error {
		for _, instr := range instructions {
			if x.ctx.Err() != nil {
				return x.ctx.Err()
			}
			notify.ScriptLine(nil, api.ScriptLine{
				Lineno: instr.lineno,
				Line:   instr.line,
			})
			if err := x.performf("строка %d: %q", instr.lineno, instr.line)(instr.work); err != nil {
				return err
			}
		}
		return nil
	})
	return
}

type instruction struct {
	work   workFunc
	line   string
	lineno int
}

type workFunc = func(worker) error

// parseScript получает список инструкций instructions из исходной строки сценария script.
// В случае ошибки в строке сценария возвращается строка с оибкой и её порядковый номер
func parseScript(script string, instructions *[]instruction) (string, int) {

	lineScanner := bufio.NewScanner(strings.NewReader(script))
	lineno := 1
	for lineScanner.Scan() {
		line := strings.TrimSpace(lineScanner.Text())
		if len(strings.TrimSpace(line)) == 0 {
			lineno++
			continue
		}

		scanner := parsec.NewScanner([]byte(line))
		node, _ := astScript.Parsewith(parserInstruction(), scanner)
		if node == nil {
			log.PrintErr("parse error: "+line, "line", lineno)
			return line, lineno
		}

		arg := func(n int) string { return node.GetChildren()[n].GetValue() }
		if x, f := scriptFunctions[node.GetName()]; f {
			*instructions = append(*instructions, instruction{
				work: func(w worker) error {
					return x.work(arg, w)
				},
				line:   line,
				lineno: lineno,
			})
		} else {
			panic(fmt.Sprintf("%s: %+v", node.GetName(), node))
		}
		lineno++
	}
	return "", -1
}

var (
	astScript = parsec.NewAST("script", 100)
)

type scriptFunc struct {
	argsParsers []interface{}
	work        func(argFunc, worker) error
}

type argFunc = func(int) string

var (
	pFloat   = parsec.Token(`[+-]?([0-9]+\.[0-9]*|\.[0-9]+|[0-9]+)`, "FLOAT")
	pMinutes = parsec.Token(`[0-9]+`, "minutes")

	scriptFunctions = map[string]scriptFunc{
		"pause": {
			argsParsers: []interface{}{pMinutes},
			work: func(f argFunc, w worker) error {
				n, _ := strconv.ParseInt(f(1), 10, 64)
				return delay(w, time.Minute*time.Duration(n), fmt.Sprintf("пауза %d минут", n))
			},
		},
		"temperature": {
			argsParsers: []interface{}{
				pFloat,
				pMinutes,
			},
			work: func(f argFunc, w worker) error {
				n, _ := strconv.ParseInt(f(2), 10, 64)
				t, _ := strconv.ParseFloat(f(1), 64)
				if err := setupTemperature(w, t); err != nil {
					return err
				}
				return delay(w, time.Minute*time.Duration(n), fmt.Sprintf("выдержка термокамеры %d минут", n))
			},
		},
		"gas": {
			argsParsers: []interface{}{
				parsec.Token(`[1-3]`, "valve"),
				pMinutes,
			},
			work: func(f argFunc, w worker) error {
				gas, _ := strconv.ParseInt(f(1), 10, 8)
				n, _ := strconv.ParseInt(f(2), 10, 64)
				if err := performWithWarn(w, func() error {
					return w.switchGas(int(gas))
				}); err != nil {
					return err
				}
				return delay(w, time.Minute*time.Duration(n), fmt.Sprintf("продувка газа %d: %d минут", gas, n))
			},
		},
		"save": {
			work: func(_ argFunc, w worker) error {
				return readSaveForLastTemperatureGas(w)
			},
		},
	}
)

func parserInstruction() parsec.Parser {
	var functionsParsers []interface{}

	for name, f := range scriptFunctions {
		var parsers []interface{}
		parsers = append(parsers, parsec.Atom(name, name+"-function-body"))
		parsers = append(parsers, f.argsParsers...)
		parsers = append(parsers, astScript.End("end-line"))
		functionsParsers = append(functionsParsers, astScript.And(name, nil, parsers...))
	}
	return astScript.OrdChoice("instruction", nil, functionsParsers...)
}
