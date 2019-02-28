package api

import (
	"github.com/fpawel/elco/internal/journal"
	"github.com/jmoiron/sqlx"
	"gopkg.in/reform.v1"
)

//go:generate go run github.com/fpawel/elco/cmd/utils/sqlstr/...

type Journal struct {
	db  *reform.DB
	dbx *sqlx.DB
}

func NewJournal(db *reform.DB, dbx *sqlx.DB) *Journal {
	return &Journal{db, dbx}
}

func (x *Journal) Years(_ struct{}, years *[]int) error {
	return x.dbx.Select(years, `SELECT DISTINCT year FROM work_info ORDER BY year ASC;`)
}

func (x *Journal) Months(r struct{ Year int }, months *[]int) error {
	return x.dbx.Select(months,
		`SELECT DISTINCT month FROM work_info WHERE year = ? ORDER BY month ASC;`,
		r.Year)
}

func (x *Journal) Days(r struct{ Year, Month int }, days *[]int) error {
	return x.dbx.Select(days,
		`SELECT DISTINCT day FROM work_info WHERE year = ? AND month = ? ORDER BY day ASC;`,
		r.Year, r.Month)
}

func (x *Journal) DayWorks(r struct{ Year, Month, Day int }, worksInfo *[]journal.WorkInfo) error {
	xs, err := x.db.SelectAllFrom(journal.WorkInfoTable, "WHERE year = ? AND month = ? AND day = ?", r.Year, r.Month, r.Day)
	if err != nil {
		return err
	}
	for _, x := range xs {
		*worksInfo = append(*worksInfo, *x.(*journal.WorkInfo))
	}
	return nil
}

func (x *Journal) EntriesOfWorkID(workID [1]int64, entries *[]journal.EntryInfo) error {
	xs, err := x.db.SelectAllFrom(journal.EntryInfoTable, "WHERE work_id = ?", workID[0])
	if err != nil {
		return err
	}
	for _, x := range xs {
		*entries = append(*entries, *x.(*journal.EntryInfo))
	}
	return nil
}

func (x *Journal) DayEntries(r struct{ Year, Month, Day int }, entries *[]journal.EntryInfo) error {
	xs, err := x.db.SelectAllFrom(journal.EntryInfoTable, `WHERE year = ? AND month = ? AND year = ?`, r.Year, r.Month, r.Day)
	if err != nil {
		return err
	}
	for _, x := range xs {
		*entries = append(*entries, *x.(*journal.EntryInfo))
	}
	return nil
}
