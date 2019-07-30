package data

import (
	"database/sql"
	"fmt"
)

type ProductField int

const (
	ProductFieldPlace ProductField = iota
	ProductFieldSerial

	ProductFieldFon20
	ProductField2Fon20
	ProductFieldSens20
	ProductFieldKSens20

	ProductFieldFonMinus20
	ProductFieldSensMinus20

	ProductFieldFon50
	ProductFieldSens50
	ProductFieldKSens50

	ProductFieldI24
	ProductFieldI35
	ProductFieldI26
	ProductFieldI17
	ProductFieldNotMeasured
	ProductFieldType
	ProductFieldPointsMethod
	ProductFieldNote
)

func (s ProductInfo) OkFieldValue(x ProductField) sql.NullBool {
	switch x {

	case ProductFieldFon20:
		return okNullFloat(s.IFPlus20, s.OKMinFon20, s.OKMaxFon20)

	case ProductField2Fon20:
		return okNullFloat(s.I13, s.OKMinFon20r, s.OKMaxFon20r)

	case ProductFieldSens20:
		return okNullFloat(s.ISPlus20, s.OKMinKSens20, s.OKMaxKSens20)

	case ProductFieldKSens20:
		return okNullFloat(s.KSens20, s.OKMinKSens20, s.OKMaxKSens20)

	case ProductFieldFonMinus20:
		return okNullFloat(s.IFMinus20, s.OKDFon20)

	case ProductFieldFon50:
		return okNullFloat(s.IFPlus50, s.OKDFon50)

	case ProductFieldSens50:
		return okNullFloat(s.ISPlus50, s.OKMinKSens50, s.OKMaxKSens50)

	case ProductFieldKSens50:
		return okNullFloat(s.KSens50, s.OKMinKSens50, s.OKMaxKSens50)

	case ProductFieldNotMeasured:
		return okNullFloat(s.NotMeasured, s.OKDNotMeasured)

	}
	return sql.NullBool{}
}

func okNullFloat(v sql.NullFloat64, args ...bool) (r sql.NullBool) {
	r.Valid = v.Valid
	r.Bool = args[0]
	for _, arg := range args[1:] {
		r.Bool = r.Bool && arg
	}
	return
}

func (s ProductInfo) FieldValue(x ProductField) interface{} {
	switch x {
	case ProductFieldPlace:
		return fmt.Sprintf("%d.%d", s.Place/8+1, s.Place%8+1)
	case ProductFieldType:
		return nullStr(s.ProductTypeName)

	case ProductFieldSerial:
		return nullInt64(s.Serial)

	case ProductFieldNote:
		if s.Note.Valid {
			return s.Note.String
		}
	case ProductFieldFon20:
		return nullFloat(s.IFPlus20)

	case ProductField2Fon20:
		return nullFloat(s.I13)

	case ProductFieldSens20:
		return nullFloat(s.ISPlus20)

	case ProductFieldKSens20:
		return nullFloat(s.KSens20)

	case ProductFieldFonMinus20:
		return nullFloat(s.IFMinus20)

	case ProductFieldSensMinus20:
		return nullFloat(s.ISMinus20)

	case ProductFieldFon50:
		return nullFloat(s.IFPlus50)

	case ProductFieldSens50:
		return nullFloat(s.ISPlus50)

	case ProductFieldKSens50:
		return nullFloat(s.KSens50)

	case ProductFieldI24:
		return nullFloat(s.I24)

	case ProductFieldI35:
		return nullFloat(s.I35)

	case ProductFieldI17:
		return nullFloat(s.I17)

	case ProductFieldI26:
		return nullFloat(s.I26)

	case ProductFieldNotMeasured:
		return nullFloat(s.NotMeasured)

	case ProductFieldPointsMethod:
		return nullInt64(s.PointsMethod)
	default:
		panic(x)
	}
	return nil
}

func nullFloat(x sql.NullFloat64) interface{} {
	if x.Valid {
		return x.Float64
	}
	return nil
}

func nullInt64(x sql.NullInt64) interface{} {
	if x.Valid {
		return x.Int64
	}
	return nil
}

func nullStr(x sql.NullString) interface{} {
	if x.Valid {
		return x.String
	}
	return nil
}
