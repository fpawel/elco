package data

import (
	"database/sql"
	"github.com/pkg/errors"
	"log"
)

type ScaleType int

const (
	Fon ScaleType = iota
	Sens
)

type Temperature float64

func (s *Product) SetCurrent(t Temperature, c ScaleType, value float64) {
	v := sql.NullFloat64{Float64: value, Valid: true}
	switch c {
	case Fon:
		switch t {
		case -20:
			s.IFMinus20 = v
			return
		case 20:
			s.IFPlus20 = v
			return
		case 50:
			s.IFPlus50 = v
			return
		}
	case Sens:
		switch t {
		case -20:
			s.ISMinus20 = v
			return
		case 20:
			s.ISPlus20 = v
			return
		case 50:
			s.ISPlus50 = v
			return
		}
	}
	log.Panicf("wrong point: %v: %v", t, c)
}

func (s ProductInfo) Current(t Temperature, c ScaleType) sql.NullFloat64 {
	switch c {
	case Fon:
		switch t {
		case -20:
			return s.IFMinus20
		case 20:
			return s.IFPlus20
		case 50:
			return s.IFPlus50
		}
	case Sens:
		switch t {
		case -20:
			return s.ISMinus20
		case 20:
			return s.ISPlus20
		case 50:
			return s.ISPlus50
		}
	}
	log.Panicf("wrong point: %v: %v", t, c)
	panic("")
}

func (s ProductInfo) CurrentValue(t Temperature, c ScaleType) (float64, error) {
	v := s.Current(t, c)
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
	if _, err := s.CurrentValue(20, Fon); err != nil {
		return nil, err
	}
	if _, err := s.CurrentValue(20, Sens); err != nil {
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
