package delphi

import (
	"math"
	"testing"
	"time"
)

func TestEncodeDateTime(t *testing.T) {

	tests := []struct {
		value string
		want  float64
	}{
		{"30.08.2019 14:03:11.000", 43707.5855551505},
		{"31.08.2019 14:03:11.000", 43708.5855551505},
		{"01.09.2019 14:03:11.000", 43709.5855551505},
		{"02.09.2019 14:03:11.000", 43710.5855551505},
		{"03.09.2019 14:03:11.000", 43711.5855551505},
		{"04.09.2019 14:03:11.000", 43712.5855551505},
		{"05.09.2019 14:03:11.000", 43713.5855551505},
		{"06.09.2019 14:03:11.000", 43714.5855551505},
		{"07.09.2019 14:03:11.000", 43715.5855551505},
		{"08.09.2019 14:03:11.000", 43716.5855551505},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			dt, err := time.Parse("02.01.2006 15:04:05.000", tt.value)
			if err != nil {
				panic(err)
			}
			got := EncodeDateTime(dt)
			d := math.Abs(EncodeDateTime(dt) - tt.want)
			if d > 0.0001 {
				t.Errorf("EncodeDateTime() = %v, want %v, %v-%v=%v", got, tt.want, got, tt.want, d)
			}
		})
	}

}

//func TestEncodeDateTime(t *testing.T) {
//
//	for y := 2000; y <= 2019; y++ {
//		for m := 1; m <= 12; m++ {
//			for d := 1; d <= 29; d++{
//				if d  == 29 {
//					if _,err := encodeDate(y,m,d); err == ErrWrongDate{
//						continue
//					}
//				}
//				for h := 0; h < 24; h++{
//					for minute := 0; minute < 60; minute++{
//						for sec :=0; sec < 60; sec++{
//							for ms :=0; ms < 1000; ms++{
//
//								dt := time.Date(y,time.Month(m),d, h, minute, sec, ms * int(time.Millisecond/time.Nanosecond), time.Local)
//								const layout = "02.01.2006 15:04:05.000"
//								t.Run(dt.Format(layout), func(t *testing.T) {
//									EncodeDateTime(dt)
//									if got := EncodeDateTime(dt); got != tt.want {
//										t.Errorf("EncodeDateTime() = %v, want %v", got, tt.want)
//									}
//								})
//
//							}
//						}
//					}
//				}
//
//
//			}
//		}
//
//
//	}
//}
