package data

import (
	"database/sql"
	"time"
)

//go:generate reform

// Party represents a row in party table.
//reform:party
type Party struct {
	PartyID         int64           `reform:"party_id,pk"`
	OldPartyID      sql.NullString  `reform:"old_party_id"`
	CreatedAt       time.Time       `reform:"created_at"`
	UpdatedAt       time.Time       `reform:"updated_at"`
	ProductTypeName string          `reform:"product_type_name"`
	Concentration1  float64         `reform:"concentration1"`
	Concentration2  float64         `reform:"concentration2"`
	Concentration3  float64         `reform:"concentration3"`
	Note            sql.NullString  `reform:"note"`
	MinFon          sql.NullFloat64 `reform:"min_fon"`
	MaxFon          sql.NullFloat64 `reform:"max_fon"`
	MaxDFon         sql.NullFloat64 `reform:"max_d_fon"`
	MinKSens20      sql.NullFloat64 `reform:"min_k_sens20"`
	MaxKSens20      sql.NullFloat64 `reform:"max_k_sens20"`
	MinKSens50      sql.NullFloat64 `reform:"min_k_sens50"`
	MaxKSens50      sql.NullFloat64 `reform:"max_k_sens50"`
	MinDTemp        sql.NullFloat64 `reform:"min_d_temp"`
	MaxDTemp        sql.NullFloat64 `reform:"max_d_temp"`
	MaxDNotMeasured sql.NullFloat64 `reform:"max_d_not_measured"`
	MaxD1           sql.NullFloat64 `reform:"max_d1"`
	MaxD2           sql.NullFloat64 `reform:"max_d2"`
	MaxD3           sql.NullFloat64 `reform:"max_d3"`
	PointsMethod    int64           `reform:"points_method"`
}
