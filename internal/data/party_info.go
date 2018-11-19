package data

import (
	"database/sql"
	"time"
)

//go:generate reform

// PartyInfo represents a row in party_info table.
//reform:party_info
type PartyInfo struct {
	PartyID         int64          `reform:"party_id,pk"`
	CreatedAt       time.Time      `reform:"created_at"`
	UpdatedAt       time.Time      `reform:"updated_at"`
	ProductTypeName string         `reform:"product_type_name"`
	Concentration1  float64        `reform:"concentration1"`
	Concentration2  float64        `reform:"concentration2"`
	Concentration3  float64        `reform:"concentration3"`
	Note            sql.NullString `reform:"note"`
	Last            bool           `reform:"last"`
}
