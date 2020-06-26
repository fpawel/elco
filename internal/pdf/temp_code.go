package pdf

import (
	"fmt"
	"github.com/fpawel/elco/internal/data"
	"math"
	"sort"
)

func tempCodes(p data.ProductInfo) (code1, code2 string) {
	code1, code2 = " ", " "
	if p.DFon50.Valid {
		code1 = fmt.Sprintf("%02d", tempCode1(p.DFon50.Float64))
	}
	if p.KSens50.Valid && p.KSensMinus20.Valid {
		code2 = fmt.Sprintf("%02d", tempCode2(p.KSensMinus20.Float64, p.KSens50.Float64))
	}
	return
}

func tempCode1(dFon50 float64) int {

	type T = struct {
		n int
		a float64
	}
	var xs []T
	for i, v := range []float64{0, 0.3, 0.6, 0.9, 1.2, 1.5, -0.3, -0.6, -0.9, -1.2, -1.5} {
		xs = append(xs, T{i + 1, v})
	}
	sort.Slice(xs, func(i, j int) bool {
		return math.Abs(dFon50-xs[i].a) < math.Abs(dFon50-xs[j].a)
	})
	return xs[0].n
}

func tempCode2(k20, k50 float64) int {

	type T struct {
		n   int
		k50 float64
	}
	var xs []T

	for i, v := range []float64{100, 110, 120, 130} {
		xs = append(xs, T{i, v})
	}
	sort.Slice(xs, func(i, j int) bool {
		return math.Abs(xs[i].k50-k50) < math.Abs(xs[j].k50-k50)
	})
	n := xs[0].n*2 + 1

	if math.Abs(k20-60) < math.Abs(k20-40) {
		n++
	}
	return n
}
