package delphi

import (
	"github.com/ansel1/merry"
	"time"
)

type GoDateTime struct {
	Year,
	Month,
	Day,
	Hour,
	Minute,
	Second,
	Millisecond int
}

func NewDateTime(t time.Time) GoDateTime {
	return GoDateTime{
		Year:        t.Year(),
		Month:       int(t.Month()),
		Day:         t.Day(),
		Hour:        t.Hour(),
		Minute:      t.Minute(),
		Second:      t.Second(),
		Millisecond: t.Nanosecond() / 1000000,
	}
}

var (
	ErrWrongTime = merry.New("wrong time")
	ErrWrongDate = merry.New("wrong date")
)

func EncodeDateTime(t time.Time) float64 {
	dt, err := encodeDate(t.Year(), int(t.Month()), t.Day())
	if err != nil {
		panic(err)
	}
	tm, err := encodeTime(t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1000000)
	if err != nil {
		panic(err)
	}
	return float64(dt) + tm
}

func encodeTime(h, m, s, ms int) (float64, error) {
	const (
		HoursPerDay = 24
		MinsPerHour = 60
		SecsPerMin  = 60
		MSecsPerSec = 1000
		MinsPerDay  = HoursPerDay * MinsPerHour
		SecsPerDay  = MinsPerDay * SecsPerMin
		//SecsPerHour = SecsPerMin * MinsPerHour
		MSecsPerDay = SecsPerDay * MSecsPerSec
	)
	if !(h < HoursPerDay && m < MinsPerHour && s < SecsPerMin && ms < MSecsPerSec) {
		return 0, ErrWrongTime
	}
	t := (h * (MinsPerHour * SecsPerMin * MSecsPerSec)) +
		(m * SecsPerMin * MSecsPerSec) +
		(s * MSecsPerSec) + ms
	if t < 0 || t >= MSecsPerDay {
		return 0, ErrWrongTime.Here().Appendf("%d:%d:%d.%d")
	}
	return float64(t) / float64(MSecsPerDay), nil

}

func encodeDate(y, m, d int) (int, error) {
	dayTable := nonLeapMonthsDays
	if (y%4 == 0) && (y%100 != 0 || y%400 == 0) {
		dayTable = leapMonthsDays
	}
	if y >= 1 && y <= 9999 && m >= 1 && m <= 12 && d >= 1 && d <= dayTable[m-1] {
		for i := 1; i < m; i++ {
			d += dayTable[i-1]
		}
		i := y - 1
		return i*365 + i/4 - i/100 + i/400 + d - dateDelta, nil
	}
	return 0, ErrWrongDate.Here().Appendf("%d/%d/%d", y, m, d)
}

var (
	leapMonthsDays    = []int{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	nonLeapMonthsDays = []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
)

//  Days between 1/1/0001 and 12/31/1899
const dateDelta = 693594
