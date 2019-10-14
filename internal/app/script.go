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
		}

		add := func(work workFunc) {
			*instructions = append(*instructions, instruction{
				work:   work,
				line:   line,
				lineno: lineno,
			})
		}

		scanner := parsec.NewScanner([]byte(line))
		node, _ := astScript.Parsewith(parserInstruction, scanner)
		if node == nil {
			log.PrintErr("parse error: "+line, "line", lineno)
			return line, lineno
		}

		arg := func() string { return node.GetChildren()[1].GetValue() }

		switch node.GetName() {
		case "TEMP":
			t, _ := strconv.ParseFloat(arg(), 64)
			add(func(w worker) error {
				return setupTemperature(w, t)
			})
			log.Println(lineno, ":", "temperature", t)
		case "GAS":
			gas, _ := strconv.ParseInt(arg(), 10, 8)
			add(func(w worker) error {
				return w.switchGas(int(gas))
			})
			log.Println("gas", gas)
		case "PAUSE":
			n, _ := strconv.ParseInt(arg(), 10, 64)
			add(func(w worker) error {
				return delay(w, time.Minute*time.Duration(n), fmt.Sprintf("пауза %d минут", n))
			})
			log.Println(lineno, ":", "pause", n)
		case "SAVE":
			add(readSaveForLastTemperatureGas)
			log.Println(lineno, ":", "save")
		default:
			panic(fmt.Sprintf("%s: %+v", node.GetName(), node))
		}
		lineno++
	}
	return "", -1
}

var (
	astScript         = parsec.NewAST("script", 100)
	parserInstruction = func() parsec.Parser {
		pFloat := parsec.Token(`[+-]?([0-9]+\.[0-9]*|\.[0-9]+|[0-9]+)`, "FLOAT")
		pEnd := astScript.End("END")
		return astScript.OrdChoice(
			"INSTR", nil,
			astScript.And("TEMP", nil,
				parsec.Atom("temperature", "TEMP-KWD"),
				pFloat,
				pEnd,
			),
			astScript.And("GAS", nil,
				parsec.Atom("gas", "GAS-KWD"),
				parsec.Token(`[1-3]`, "GAS-VALVE"),
				pEnd,
			),
			astScript.And("PAUSE", nil,
				parsec.Atom("pause", "PAUSE-KWD"),
				parsec.Token(`[0-9]+`, "PAUSE-VALUE"),
				pEnd,
			),
			astScript.And("SAVE", nil,
				parsec.Atom("save", "SAVE-KWD"),
				pEnd,
			),
		)
	}()
)
