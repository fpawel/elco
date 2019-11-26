package chipmem

import (
	"fmt"
	"github.com/ansel1/merry"
	"math"
	"strings"
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

func tempPoints(values []string, fonM TableXY, sensM TableXY) error {
	if len(values)%3 != 0 {
		return merry.New("sequence length is not a multiple of three")
	}

	for n := 0; n < len(values); n += 3 {
		strT := strings.TrimSpace(values[n+0])
		if len(strT) == 0 {
			continue
		}

		t, err := parseFloat(values[n])
		if err != nil {
			return merry.Appendf(err, "строка %d", n)
		}
		strI := strings.TrimSpace(values[n+1])
		if len(strI) > 0 {
			var i float64
			i, err = parseFloat(strI)
			if err != nil {
				return merry.Appendf(err, "строка %d", n)
			}
			fonM[t] = i
		}
		strS := strings.TrimSpace(values[n+2])
		if len(strS) > 0 {
			var k float64
			k, err = parseFloat(strS)
			if err != nil {
				return merry.Appendf(err, "строка %d", n)
			}
			sensM[t] = k
		}
	}
	// r = data.NewTempPoints(fonM, sensM)
	return nil
}

func tempValues(r TempPoints) (xs []string) {
	var mainTemps = map[float64]struct{}{
		-50: {},
		-40: {},
		-30: {},
		-20: {},
		-10: {},
		-5:  {},
		0:   {},
		5:   {},
		10:  {},
		20:  {},
		30:  {},
		40:  {},
		50:  {},
		60:  {},
	}
	for i, t := range r.Temp {
		if _, f := mainTemps[t]; f {
			delete(mainTemps, t)
			xs = append(xs,
				fmt.Sprintf("%v", t),
				fmt.Sprintf("%v", r.Fon[i]),
				fmt.Sprintf("%v", r.Sens[i]))
		}
	}
	return
}
