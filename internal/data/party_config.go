package data

import (
	"database/sql"
	"github.com/fpawel/goutils"
	"github.com/pkg/errors"
	"gopkg.in/reform.v1"
	"strconv"
	"strings"
)

func SetConfigValue(db *reform.DB, property, value string) (err error) {

	var party Party
	party, err = GetLastParty(db)

	switch property {

	case "ProductType":
		party.ProductTypeName = value
		return db.Save(&party)

	case "Gas1":
		party.Concentration1, err = goutils.ParseFloat(value)
		if err == nil {
			err = db.Save(&party)
		}
		return

	case "Gas2":
		party.Concentration2, err = goutils.ParseFloat(value)
		if err == nil {
			err = db.Save(&party)
		}
		return

	case "Gas3":
		party.Concentration3, err = goutils.ParseFloat(value)
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
				if v.Float64, err = goutils.ParseFloat(value); err != nil {
					return err
				}
				v.Valid = true
			}
			f()
			return db.Save(&party)
		}
	}
	return errors.Errorf("%q: wrong party property")
}
