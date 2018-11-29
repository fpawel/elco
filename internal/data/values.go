package data

import (
	"database/sql"
	"github.com/pkg/errors"
)

type CurrScaleType int

const (
	Fon CurrScaleType = iota
	Sens
)

func (s ProductInfo) Current(c CurrScaleType, t float64) sql.NullFloat64 {
	switch c {
	case Fon:
		switch t {
		case -20:
			return s.IFMinus20
		case 20:
			return s.IFPlus20
		case 50:
			return s.IFPlus50
		default:
			panic("wrong temperature")
		}
	case Sens:
		switch t {
		case -20:
			return s.ISMinus20
		case 20:
			return s.ISPlus20
		case 50:
			return s.ISPlus50
		default:
			panic("wrong temperature")
		}
	default:
		panic("wrong scale point")
	}
}

func (s ProductInfo) CurrentValue(c CurrScaleType, t float64) (float64, error) {
	v := s.Current(c, t)
	if !v.Valid {
		str := "фонового тока"
		if c == Sens {
			str = "тока чувствительности"
		}
		return 0, errors.Errorf("нет значения %s при %g⁰С", str, t)
	}
	return v.Float64, nil
}

func (s ProductInfo) KSensPercentValues(includeMinus20 bool) (map[float64]float64, error) {
	if _, err := s.CurrentValue(Fon, 20); err != nil {
		return nil, err
	}
	if _, err := s.CurrentValue(Sens, 20); err != nil {
		return nil, err
	}
	if !s.KSens50.Valid {
		return nil, errors.New("нет значения к-та чувствительности при 50⁰С")
	}

	r := map[float64]float64{
		20: 100,
		50: s.KSens50.Float64,
	}
	if s.KSensMinus20.Valid {
		r[-20] = s.KSensMinus20.Float64
	} else {
		if includeMinus20 {
			return nil, errors.New("нет значения к-та чувствительности при -20⁰С")
		}
	}
	return r, nil
}
