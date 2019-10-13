package internal

import (
	"bufio"
	"fmt"
	parsec "github.com/prataprc/goparsec"
	"strings"
	"testing"
)

func TestScript(t *testing.T) {

	//var script []func()
	ast := parsec.NewAST("one", 100)

	pInt := parsec.Token(`[0-9]+`, "INT")
	pFloat := parsec.Token(`[+-]?([0-9]+\.[0-9]*|\.[0-9]+|[0-9]+)`, "FLOAT")

	type Q = parsec.Queryable
	wh := func(f func(parsec.Queryable)) parsec.ASTNodify {
		return func(name string, s parsec.Scanner, node parsec.Queryable) parsec.Queryable {
			f(node)
			return node
		}
	}
	pEnd := ast.End("END")

	pInstr := ast.OrdChoice(
		"INSTR", nil,
		ast.And("TEMP",
			func(name string, s parsec.Scanner, node parsec.Queryable) parsec.Queryable {
				return node
			},
			parsec.Atom("temperature", "TEMP-KWD"),
			pFloat,
			pEnd,
		),
		ast.And("GAS",
			wh(func(node Q) {

			}),
			parsec.Atom("gas", "GAS-KWD"),
			pInt,
			pEnd,
		),
		ast.And("PAUSE",
			wh(func(node Q) {

			}),
			parsec.Atom("pause", "PAUSE-KWD"),
			pInt,
			pEnd,
		),
		ast.And("SAVE",
			wh(func(node Q) {

			}),
			parsec.Atom("save", "SAVE-KWD"),
			pEnd,
		),
	)

	s := `save
pause 1222
temperature 20
gas 178 dfg
gas 1   
pause 12`

	lineScanner := bufio.NewScanner(strings.NewReader(s))
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
			break
		}
		fmt.Println(str, ":")
		ast.Prettyprint()
		lineNo++
	}
}
