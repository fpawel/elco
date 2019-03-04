package intrng

import (
	"fmt"
	"github.com/prataprc/goparsec"
	"regexp"
	"strconv"
)

func Format(b []int) (result string) {
	IterateOverRanges(b, func(x,y int) {
		if result != "" {
			result += " "
		}
		if x == y {
			result += fmt.Sprintf("%d", x)
		} else {
			result += fmt.Sprintf("%d-%d", x, y)
		}
	})
	return
}

func IterateOverRanges(b []int, f func (int, int))  {
	for i := 0; i < len(b); i++ {
		if i < len(b)-1 && (b[i] == b[i+1] || b[i]+1 == b[i+1]) {
			continue
		}

		x, y := b[i], b[i]
		for j := i; j > -1; j-- {
			y = b[j]
			if j > 0 && !(b[j-1] == b[j] || b[j-1]+1 == b[j]) {
				break
			}
		}
		f(y, x)
	}
	return
}

func vector2scalar(nodes []parsec.ParsecNode) parsec.ParsecNode {
	return nodes[0]
}

func vector2number(nodes []parsec.ParsecNode) parsec.ParsecNode {
	switch node := nodes[0].(type) {
	case *parsec.Terminal:
		n, err := strconv.Atoi(node.Value)
		if err != nil {
			panic(err)
		}
		return n

	default:
		panic("unknown node type")
	}
}

type numberRange struct {
	x, y int
}

func vector2numberRange(nodes []parsec.ParsecNode) parsec.ParsecNode {
	re := regexp.MustCompile(`(\d+)\s*-\s*(\d+)`)
	node := nodes[0]
	switch node := node.(type) {
	case *parsec.Terminal:
		xs := re.FindAllStringSubmatch(node.Value, -1)
		if len(xs) != 1 {
			panic("")
		}
		if len(xs[0]) != 3 {
			panic("")
		}
		var r numberRange
		var err error
		r.x, err = strconv.Atoi(xs[0][1])
		if err != nil {
			panic(err)
		}
		r.y, err = strconv.Atoi(xs[0][2])
		if err != nil {
			panic(err)
		}
		return r
	default:
		panic(node)
	}
}

func Parse(str string) (r []int) {

	number := parsec.OrdChoice(vector2number, parsec.Token(`\d+`, "NUMBER"))
	numbers := parsec.OrdChoice(vector2numberRange, parsec.Token(`\d+\s*-\s*\d+`, "NUMBERS"))
	item := parsec.OrdChoice(vector2scalar, numbers, number)
	scanner := parsec.NewScanner([]byte(str))
	var node parsec.ParsecNode
	for node, scanner = item(scanner); node != nil; node, scanner = item(scanner) {
		switch node := node.(type) {
		case int:
			r = append(r, node)
		case numberRange:
			for v := node.x; v <= node.y; v++ {
				r = append(r, v)
			}
		default:
			panic(node)
		}
		_, scanner = scanner.SkipWS()
	}

	return
}
