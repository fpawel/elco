package firmware

import (
	"fmt"
	"github.com/fpawel/elco/internal/data"
	"math"
)

type M = map[float64]float64

const (
	fon = data.Fon
	//sens = data.Sens
)

func srcFon2(p data.ProductInfo) (y M, err error) {
	y = M{}

	y[20], err = p.CurrentValue(fon, 20)
	if err != nil {
		return
	}
	y[50], err = p.CurrentValue(fon, 50)
	if err != nil {
		return
	}

	y[40] = (y[50]-y[20])*0.5 + y[20]
	y[-40] = 0
	y[-20] = y[20] * 0.2
	y[0] = y[20] * 0.5
	y[30] = (y[40]-y[20])*0.5 + y[20]
	y[45] = (y[50]-y[40])*0.5 + y[40]

	return
}

func srcFon3(p data.ProductInfo) (y M, err error) {
	y = M{}
	y[-20], err = p.CurrentValue(fon, -20)
	if err != nil {
		return
	}
	y[20], err = p.CurrentValue(fon, 20)
	if err != nil {
		return
	}
	y[50], err = p.CurrentValue(fon, 50)
	if err != nil {
		return
	}
	y[40] = (y[50]-y[20])*0.5 + y[20]
	y[-40] = y[-20] - 0.5*(y[20]-y[-20])
	y[0] = y[20] - 0.5*(y[20]-y[-20])
	y[30] = (y[40]-y[20])*0.5 + y[20]
	y[45] = (y[50]-y[40])*0.5 + y[40]
	return
}

func srcSens2(p data.ProductInfo) (M, error) {
	y, err := p.KSensPercentValues(false)
	if err == nil {
		y[40] = (y[50]-y[20])*0.5 + y[20]
		y[-40] = 30
		y[-20] = 58
		y[0] = 82
		y[30] = (y[40]-y[20])*0.5 + y[20]
		y[45] = (y[50]-y[40])*0.5 + y[40]
	}
	return y, err
}

func srcSens3(p data.ProductInfo) (M, error) {
	y, err := p.KSensPercentValues(true)
	if err == nil {
		//if y[-20] > 0 && y[-20] < 0.45*y[20] {
		//	return y, errors.Errorf(
		//		"ток чувствительности: I(-20)=%v, I(+20)=%v, I(-20)>0, I(-20)<0.45*I(+20)",
		//		y[-20], y[20])
		//}
		y[0] = (y[20]-y[-20])*0.5 + y[-20]
		y[40] = y[50] - y[20]*0.5 + y[20]
		y[45] = y[50] - y[40]*0.5 + y[40]
		y[30] = y[40] - y[20]*0.5 + y[20]
		y[-40] = 2*y[-20] - y[0]
		if y[-20] > 0 {
			y[-40] += 1.2 * (45 - y[-20]) / (0.43429 * math.Log(y[-20]))
		}
	}
	return y, err
}

func srcSens(p data.ProductInfo) (M, error) {
	switch p.PointsMethod {
	case 2:
		return srcSens2(p)
	case 3:
		return srcSens3(p)
	default:
		panic(fmt.Sprintf("wrong points method: %d", p.PointsMethod))
	}
}

func srcFon(p data.ProductInfo) (M, error) {
	switch p.PointsMethod {
	case 2:
		return srcFon2(p)
	case 3:
		return srcFon3(p)
	default:
		panic(fmt.Sprintf("wrong points method: %d", p.PointsMethod))
	}
}
