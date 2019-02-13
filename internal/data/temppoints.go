package data

import (
	"math"
)

type TempPoints struct {
	Temp, Fon, Sens [250]float64
}

func NewTempPoints(fonM, sensM TableXY) (r TempPoints) {
	minusOne := func(_ float64) float64 {
		return -1
	}
	fFon := minusOne
	fSens := minusOne
	if len(fonM) > 0 {
		fFon = NewApproximationTable(fonM).F
	}
	if len(sensM) > 0 {
		fSens = NewApproximationTable(sensM).F
	}
	i := 0
	for t := float64(-124); t <= 125; t++ {
		r.Temp[i] = t
		r.Fon[i] = math.Round(fFon(t))
		r.Sens[i] = math.Round(fSens(t))
		i++
	}
	return
}
