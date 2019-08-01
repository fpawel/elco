package api

import (
	"database/sql"
	"github.com/fpawel/elco/internal/data"
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
	ProductTypeName string
	Concentration1  float64
	Concentration2  float64
	Concentration3  float64
	Note            sql.NullString
	MinFon          sql.NullFloat64
	MaxFon          sql.NullFloat64
	MaxDFon         sql.NullFloat64
	MinKSens20      sql.NullFloat64
	MaxKSens20      sql.NullFloat64
	MinKSens50      sql.NullFloat64
	MaxKSens50      sql.NullFloat64
	MinDTemp        sql.NullFloat64
	MaxDTemp        sql.NullFloat64
	MaxDNotMeasured sql.NullFloat64
	PointsMethod    int64
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
