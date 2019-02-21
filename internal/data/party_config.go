package data

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/reform.v1"
	"strconv"
	"strings"
)

func PartyConfigProperties(db *reform.DB) ([]ConfigProperty, error) {
	var party Party
	if err := GetLastParty(db, &party); err != nil {
		return nil, err
	}

	productTypesNames, err := ListProductTypeNames(db)
	if err != nil {
		return nil, err
	}

	f := func(v sql.NullFloat64) string {
		if v.Valid {
			return fmt.Sprintf("%v", v.Float64)
		}
		return ""
	}

	return []ConfigProperty{
		{
			Hint:      "Исполнение",
			Name:      "ProductType",
			ValueType: VtString,
			Value:     party.ProductTypeName,
			List:      productTypesNames,
		},
		{
			Hint:      "ПГС1",
			Name:      "Gas1",
			ValueType: VtFloat,
			Value:     fmt.Sprintf("%v", party.Concentration1),
		},
		{
			Hint:      "ПГС2",
			Name:      "Gas2",
			ValueType: VtFloat,
			Value:     fmt.Sprintf("%v", party.Concentration2),
		},
		{
			Hint:      "ПГС3",
			Name:      "Gas3",
			ValueType: VtFloat,
			Value:     fmt.Sprintf("%v", party.Concentration3),
		},
		{
			Hint:      "Примечание",
			Name:      "Note",
			ValueType: VtString,
			Value:     fmt.Sprintf("%v", party.Note.String),
		},

		{
			Hint:      "Кол-во точек для расчёта",
			Name:      "PointsMethod",
			ValueType: VtString,
			Value:     strconv.Itoa(int(party.PointsMethod)),
			List:      []string{"2", "3"},
		},

		{
			Hint:      "Фон.мин, мкА",
			Name:      "MinFon",
			ValueType: VtNullFloat,
			Value:     f(party.MinFon),
		},
		{
			Hint:      "Фон.мax, мкА",
			Name:      "MaxFon",
			ValueType: VtNullFloat,
			Value:     f(party.MaxFon),
		},
		{
			Hint:      "D.фон.мax, мкА",
			Name:      "MaxDFon",
			ValueType: VtNullFloat,
			Value:     f(party.MaxDFon),
		},
		{
			Hint:      "Кч20.мин, мкА/мг/м3",
			Name:      "MinKSens20",
			ValueType: VtNullFloat,
			Value:     f(party.MinKSens20),
		},
		{
			Hint:      "Кч20.макс, мкА/мг/м3",
			Name:      "MaxKSens20",
			ValueType: VtNullFloat,
			Value:     f(party.MaxKSens20),
		},
		{
			Hint:      "Кч50.мин, мкА/мг/м3",
			Name:      "MinKSens50",
			ValueType: VtNullFloat,
			Value:     f(party.MinKSens50),
		},
		{
			Hint:      "Кч50.макс, мкА/мг/м3",
			Name:      "MaxKSens50",
			ValueType: VtNullFloat,
			Value:     f(party.MaxKSens50),
		},
		{
			Hint:      "Dt.мин, мкА",
			Name:      "MinDTemp",
			ValueType: VtNullFloat,
			Value:     f(party.MinDTemp),
		},
		{
			Hint:      "Dt.мин, мкА",
			Name:      "MaxDTemp",
			ValueType: VtNullFloat,
			Value:     f(party.MaxDTemp),
		},
		{
			Hint:      "Dn.макс, мкА",
			Name:      "MaxDNotMeasured",
			ValueType: VtNullFloat,
			Value:     f(party.MaxDNotMeasured),
		},
	}, nil
}

func SetPartyConfigValue(db *reform.DB, property, value string) (err error) {

	var party Party
	if err := GetLastParty(db, &party); err != nil {
		return err
	}

	switch property {

	case "ProductType":
		party.ProductTypeName = value
		return db.Save(&party)

	case "Gas1":
		party.Concentration1, err = parseFloat(value)
		if err == nil {
			err = db.Save(&party)
		}
		return

	case "Gas2":
		party.Concentration2, err = parseFloat(value)
		if err == nil {
			err = db.Save(&party)
		}
		return

	case "Gas3":
		party.Concentration3, err = parseFloat(value)
		if err == nil {
			err = db.Save(&party)
		}
		return

	case "Note":
		party.Note.String = strings.TrimSpace(value)
		party.Note.Valid = len(party.Note.String) > 0
		err = db.Save(&party)
		return

	case "PointsMethod":
		party.PointsMethod, err = strconv.ParseInt(value, 10, 8)
		if err != nil {
			return err
		}
		err = db.Save(&party)
		return

	default:
		var v sql.NullFloat64
		fs := map[string]func(){
			"MinFon": func() {
				party.MinFon = v
			},
			"MaxFon": func() {
				party.MaxFon = v
			},
			"MaxDFon": func() {
				party.MaxDFon = v
			},
			"MinKSens20": func() {
				party.MinKSens20 = v
			},
			"MaxKSens20": func() {
				party.MaxKSens20 = v
			},
			"MinKSens50": func() {
				party.MinKSens50 = v
			},
			"MaxKSens50": func() {
				party.MaxKSens50 = v
			},
			"MinDTemp": func() {
				party.MinDTemp = v
			},
			"MaxDTemp": func() {
				party.MaxDTemp = v
			},
			"MaxDNotMeasured": func() {
				party.MaxDNotMeasured = v
			},
		}
		if f, ok := fs[property]; ok {
			if len(strings.TrimSpace(value)) > 0 {
				if v.Float64, err = parseFloat(value); err != nil {
					return err
				}
				v.Valid = true
			}
			f()
			return db.Save(&party)
		}
	}
	return errors.Errorf("%q: wrong party property", property)
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.Replace(s, ",", ".", -1), 64)
}
