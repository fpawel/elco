package internal

import (
	"bufio"
	"fmt"
	parsec "github.com/prataprc/goparsec"
	"strings"
	"testing"
)

type Instruction struct {
	Name  string
	Param int
}

func TestScript(t *testing.T) {

	//var script []func()
	ast := parsec.NewAST("one", 100)

	pInt := parsec.Token(`[0-9]+`, "INT")
	pFloat := parsec.Token(`[+-]?([0-9]+\.[0-9]*|\.[0-9]+|[0-9]+)`, "FLOAT")

	pInstruction := ast.OrdChoice(
		"INSTRUCTION", nil,
		ast.And("TEMPERATURE",
			func(name string, s parsec.Scanner, node parsec.Queryable) parsec.Queryable {
				return node
			},
			parsec.Atom("temperature", "TEMPERATURE-KEYWORD"),
			pFloat,
		),
		ast.And("GAS",
			func(name string, s parsec.Scanner, node parsec.Queryable) parsec.Queryable {
				return node
			},
			parsec.Atom("gas", "GAS-KEYWORD"),
			pInt,
		),
		ast.And("PAUSE",
			func(name string, s parsec.Scanner, node parsec.Queryable) parsec.Queryable {
				return node
			},
			parsec.Atom("pause", "PAUSE-KEYWORD"),
			pInt,
		),
		ast.And("SAVE",
			func(name string, s parsec.Scanner, node parsec.Queryable) parsec.Queryable {
				return node
			},
			parsec.Atom("save", "SAVE-KEYWORD"),
		),
	)
	pLine := ast.And("INSTRUCTION-LINE", nil, pInstruction, ast.End("END"))

	s := `
save
pause -1222
temperature 20
gas 178
gas 1  bubu
pause 12
`

	lineScanner := bufio.NewScanner(strings.NewReader(s))
	lineNo := 1
	for lineScanner.Scan() {
		str := strings.TrimSpace(lineScanner.Text())
		if len(str) == 0 {
			lineNo++
			continue
		}
		ast.Reset()
		scanner := parsec.NewScanner([]byte(str))
		node, _ := ast.Parsewith(pLine, scanner)
		if node == nil {
			fmt.Printf("not parsed: line %d: %q\n", lineNo, str)
			break
		}
		for _, node := range node.GetChildren() {
			xs := node.GetChildren()
			if len(xs) == 0 {
				continue
			}
			if len(xs) == 1 {
				fmt.Println(xs[0].GetValue())
				continue
			}
			fmt.Println(xs[0].GetValue(), xs[1].GetValue())
		}
		lineNo++
		//fmt.Println(str , ":")
		//ast.Prettyprint()
	}
}

func setTemperature(t float64) {
	fmt.Println("set temperature", t)
}
func setGas(gas int) {
	fmt.Println("set gas", gas)
}
func pause(n int) {
	fmt.Println("pause", n)
}
func save() {
	fmt.Println("save")
}
