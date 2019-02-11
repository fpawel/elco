package data

import "sort"

type TableXY = map[float64]float64

// ApproximationTable - таблица кусочно-линейной апроксимации
type ApproximationTable struct {
	vx, vy []float64
}

func NewApproximationTable(xy TableXY) *ApproximationTable {
	n := len(xy)
	if n == 0 {
		panic("map must be not empty")
	}
	tbl := new(ApproximationTable)

	var xys [][2]float64
	for x, y := range xy {
		xys = append(xys, [2]float64{x, y})
	}
	sort.Slice(xys, func(i, j int) bool {
		return xys[i][0] < xys[j][0]
	})
	for _, a := range xys {
		tbl.vx = append(tbl.vx, a[0])
		tbl.vy = append(tbl.vy, a[1])
	}
	return tbl
}

func (tbl *ApproximationTable) F(x float64) float64 {
	if x < tbl.vx[0] {
		return tbl.vy[0]
	}
	for i := 1; i < len(tbl.vx); i++ {
		if tbl.vx[i-1] <= x && x < tbl.vx[i] {
			b := (tbl.vy[i] - tbl.vy[i-1]) / (tbl.vx[i] - tbl.vx[i-1])
			a := tbl.vy[i-1] - b*tbl.vx[i-1]
			return a + b*x
		}
	}
	return tbl.vy[len(tbl.vy)-1]
}
