package firmware

import (
	"database/sql"
	"github.com/fpawel/elco/internal/data"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"math"
	"sort"
)

type M = map[float64]float64

type DeltaCurrToKSensPercentFunc = func(curFon, curSens float64) float64

type CalcMethod int

const (
	CalcMethod2 = 2
	CalcMethod3 = 3
)

func srcFon2(p data.ProductInfo) (y M, err error) {
	if !p.IFPlus20.Valid {
		err = multierror.Append(err, errors.New("нет значения фонового тока при +20\""))
	}
	if !p.IFPlus50.Valid {
		err = multierror.Append(err, errors.New("нет значения фонового тока при +50\""))
	}

	if err == nil {
		y = M{}
		y[20] = p.IFPlus20.Float64
		y[50] = p.IFPlus50.Float64
		y[40] = (y[50]-y[20])*0.5 + y[20]
		y[-40] = 0
		y[-20] = y[20] * 0.2
		y[0] = y[20] * 0.5
		y[30] = (y[40]-y[20])*0.5 + y[20]
		y[45] = (y[50]-y[40])*0.5 + y[40]
	}
	return
}

func srcFon3(p data.ProductInfo) (y M, err error) {
	if !p.IFMinus20.Valid {
		err = multierror.Append(err, errors.New("нет значения фонового тока при -20\""))
	}
	if !p.IFPlus20.Valid {
		err = multierror.Append(err, errors.New("нет значения фонового тока при +20\""))
	}
	if !p.IFPlus50.Valid {
		err = multierror.Append(err, errors.New("нет значения фонового тока при +50\""))
	}
	if err == nil {
		y = M{}
		y[-20] = p.IFMinus20.Float64
		y[20] = p.IFPlus20.Float64
		y[50] = p.IFPlus50.Float64
		y[40] = (y[50]-y[20])*0.5 + y[20]
		y[-40] = y[-20] - 0.5*(y[20]-y[-20])
		y[0] = y[20] - 0.5*(y[20]-y[-20])
		y[30] = (y[40]-y[20])*0.5 + y[20]
		y[45] = (y[50]-y[40])*0.5 + y[40]
	}
	return
}

func srcSens2(p data.ProductInfo) (y M, err error) {

	if !p.ISPlus20.Valid {
		err = multierror.Append(err, errors.New("нет значения тока чувствительности при +20\""))
	}
	if !p.ISPlus50.Valid {
		err = multierror.Append(err, errors.New("нет значения тока чувствительности при +50\""))
	}

	if err == nil {
		y = M{}
		y[20] = p.ISPlus20.Float64
		y[50] = p.ISPlus50.Float64
		y[40] = (y[50]-y[20])*0.5 + y[20]
		y[-40] = 30
		y[-20] = 58
		y[0] = 82
		y[30] = (y[40]-y[20])*0.5 + y[20]
		y[45] = (y[50]-y[40])*0.5 + y[40]
	}
	return
}

func srcSens3(p data.ProductInfo) (y M, err error) {
	if !p.ISMinus20.Valid {
		err = multierror.Append(err, errors.New("нет значения тока чувствительности при -20\""))
	}
	if !p.ISPlus20.Valid {
		err = multierror.Append(err, errors.New("нет значения тока чувствительности при +20\""))
	}
	if !p.ISPlus50.Valid {
		err = multierror.Append(err, errors.New("нет значения тока чувствительности при +50\""))
	}

	if err == nil {
		y = M{}
		y[-20] = p.ISMinus20.Float64
		y[20] = p.ISPlus20.Float64
		y[50] = p.ISPlus50.Float64

		if y[-20] > 0 && y[-20] < 0.45*y[20] {
			err = multierror.Append(err, errors.Errorf(
				"ток чувствительности: I(-20)=%v, I(+20)=%v, I(-20)>0, I(-20)<0.45*I(+20)",
				y[-20], y[20]))
		} else {
			y[0] = (y[20]-y[-20])*0.5 + y[-20]
			y[40] = y[50] - y[20]*0.5 + y[20]
			y[45] = y[50] - y[40]*0.5 + y[40]
			y[30] = y[40] - y[20]*0.5 + y[20]
			y[-40] = 2*y[-20] - y[0]
			if y[-20] > 0 {
				y[-40] += 1.2 * (45 - y[-20]) / (0.43429 * math.Log(y[-20]))
			}
		}
	}
	return
}

func srcSens(p data.ProductInfo, m CalcMethod) (M, error) {
	switch m {
	case CalcMethod2:
		return srcSens2(p)
	case CalcMethod3:
		return srcSens3(p)
	default:
		panic(m)
	}
}

func srcFon(p data.ProductInfo, m CalcMethod) (M, error) {
	switch m {
	case CalcMethod2:
		return srcFon2(p)
	case CalcMethod3:
		return srcFon3(p)
	default:
		panic(m)
	}
}

// кусочно-линейная апроксимация
func pieceWiseLinearApproximation(xy M, x float64) float64 {

	n := len(xy)
	if n == 0 {
		panic("map must be not empty")
	}

	var xs, ys []float64
	{
		var xys [][2]float64
		for x, y := range xy {
			xys = append(xys, [2]float64{x, y})
		}
		sort.Slice(xys, func(i, j int) bool {
			return xys[i][0] < xys[j][0]
		})
		for _, a := range xys {
			xs = append(xs, a[0])
			ys = append(ys, a[1])
		}
	}
	if x < xs[0] {
		return ys[0]
	}
	for i := 1; i < n; i++ {
		if xs[i-1] <= x && x < xs[i] {
			b := (ys[i] - ys[i-1]) / (xs[i] - xs[i-1])
			a := ys[i-1] - b*xs[i-1]
			return a + b*x
		}
	}
	return ys[n-1]
}

func productTemperatureToKSensPercent(x data.ProductInfo) M {
	if !x.IFPlus20.Valid || !x.ISPlus20.Valid {
		return nil
	}
	d := x.ISPlus20.Float64 - x.IFPlus20.Float64
	if d == 0 {
		return nil
	}
	deltaCurrToKSensPercent := func(curFon, curSens float64) float64 {
		return 100 * math.Abs((curSens-curFon)/d)
	}
	r := make(map[float64]float64)
	ts := []float64{-20, 20, 50}
	cs := []sql.NullFloat64{x.ISMinus20, x.ISPlus20, x.ISPlus50}
	cf := []sql.NullFloat64{x.IFMinus20, x.IFPlus20, x.IFPlus50}
	for i := 0; i < 2; i++ {
		if cs[i].Valid && cf[i].Valid {
			r[ts[i]] = deltaCurrToKSensPercent(cf[i].Float64, cs[i].Float64)
		}
	}
	return r
}
