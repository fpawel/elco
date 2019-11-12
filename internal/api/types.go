package api

import (
	"database/sql"
	"fmt"
	"github.com/fpawel/elco/internal/data"
	"strconv"
	"strings"
	"time"
)

type WorkResult struct {
	WorkName string
	Tag      WorkResultTag
	Message  string
}

type WorkResultTag int

type ReadCurrent struct {
	Values []float64
	Block  int
}

type DelayInfo struct {
	TotalSeconds, ElapsedSeconds int
	What                         string
}

type GetCheckBlocksArg struct {
	Check []bool
}

type Ktx500Info struct {
	Temperature, Destination float64
	TemperatureOn, CoolOn    bool
}

type YearMonth struct {
	Year  int `db:"year"`
	Month int `db:"month"`
}

type Party1 struct {
	PartyID   int64 `db:"party_id"`
	CreatedAt time.Time
	Products  []data.ProductInfo
}

type Party2 struct {
	PartyID         int64          `db:"party_id"`
	Day             int            `db:"day"`
	ProductTypeName string         `db:"product_type_name"`
	Note            sql.NullString `db:"note"`
	Last            bool           `db:"last"`
}

type Party3 struct {
	ProductTypeName string  `yaml:"product_type_name"`
	Concentration1  float64 `yaml:"concentration1"`
	Concentration2  float64 `yaml:"concentration2"`
	Concentration3  float64 `yaml:"concentration3"`
	Note            string  `yaml:"name"`
	MinFon          string  `yaml:"min_fom"`
	MaxFon          string  `yaml:"max_fon"`
	MaxDFon         string  `yaml:"max_d_fon"`
	MinKSens20      string  `yaml:"min_k_sens20"`
	MaxKSens20      string  `yaml:"max_k_sens20"`
	MinKSens50      string  `yaml:"min_k_sens50"`
	MaxKSens50      string  `yaml:"max_k_sens50"`
	MinDTemp        string  `yaml:"min_d_temp"`
	MaxDTemp        string  `yaml:"max_d_temp"`
	MaxDNotMeasured string  `yaml:"max_d_not_measured"`
	PointsMethod    int64   `yaml:"points_method"`
	MaxD1           string  `yaml:"max_d1"`
	MaxD2           string  `yaml:"max_d2"`
	MaxD3           string  `yaml:"max_d3"`
}

func newParty3(x data.Party) (p Party3) {
	p.ProductTypeName = x.ProductTypeName
	p.Concentration1 = x.Concentration1
	p.Concentration2 = x.Concentration2
	p.Concentration3 = x.Concentration3
	p.Note = x.Note.String

	p.MinFon = formatNullFloat(x.MinFon)
	p.MaxFon = formatNullFloat(x.MaxFon)
	p.MaxDFon = formatNullFloat(x.MaxDFon)
	p.MinKSens20 = formatNullFloat(x.MinKSens20)
	p.MaxKSens20 = formatNullFloat(x.MaxKSens20)
	p.MinKSens50 = formatNullFloat(x.MinKSens50)
	p.MaxKSens50 = formatNullFloat(x.MaxKSens50)
	p.MinDTemp = formatNullFloat(x.MinDTemp)
	p.MaxDTemp = formatNullFloat(x.MaxDTemp)
	p.MaxDNotMeasured = formatNullFloat(x.MaxDNotMeasured)
	p.PointsMethod = x.PointsMethod
	p.MaxD1 = formatNullFloat(x.MaxD1)
	p.MaxD2 = formatNullFloat(x.MaxD2)
	p.MaxD3 = formatNullFloat(x.MaxD3)
	return
}

func (x Party3) SetupDataParty(p *data.Party) {
	p.ProductTypeName = x.ProductTypeName
	p.Concentration1 = x.Concentration1
	p.Concentration2 = x.Concentration2
	p.Concentration3 = x.Concentration3
	p.Note.String = strings.TrimSpace(x.Note)
	p.Note.Valid = len(p.Note.String) > 0

	p.MinFon = parseNullFloat(x.MinFon)
	p.MaxFon = parseNullFloat(x.MaxFon)
	p.MaxDFon = parseNullFloat(x.MaxDFon)
	p.MinKSens20 = parseNullFloat(x.MinKSens20)
	p.MaxKSens20 = parseNullFloat(x.MaxKSens20)
	p.MinKSens50 = parseNullFloat(x.MinKSens50)
	p.MaxKSens50 = parseNullFloat(x.MaxKSens50)
	p.MinDTemp = parseNullFloat(x.MinDTemp)
	p.MaxDTemp = parseNullFloat(x.MaxDTemp)
	p.MaxDNotMeasured = parseNullFloat(x.MaxDNotMeasured)
	p.PointsMethod = x.PointsMethod
	p.MaxD1 = parseNullFloat(x.MaxD1)
	p.MaxD2 = parseNullFloat(x.MaxD2)
	p.MaxD3 = parseNullFloat(x.MaxD3)
}

type ScriptLine struct {
	Lineno int
	Line   string
}

func newParty1(partyID int64) (r Party1) {
	p := data.GetParty(partyID)
	r.CreatedAt = p.CreatedAt
	r.PartyID = p.PartyID
	r.Products = data.ProductsInfoAll(partyID)
	return
}

func LastParty1() Party1 {
	return newParty1(data.LastPartyID())
}

func parseNullFloat(x string) sql.NullFloat64 {
	if v, err := strconv.ParseFloat(x, 64); err == nil {
		return sql.NullFloat64{
			Float64: v,
			Valid:   true,
		}
	}
	return sql.NullFloat64{}
}

func formatNullFloat(x sql.NullFloat64) string {
	if x.Valid {
		return fmt.Sprintf("%v", x.Float64)
	}
	return ""
}
