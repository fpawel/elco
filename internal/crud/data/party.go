package data

import (
	"time"
)

//go:generate reform

// Party represents a row in party table.
//reform:party
type Party struct {
	PartyID         int64     `reform:"party_id,pk"`
	OldPartyID      *string   `reform:"old_party_id"`
	CreatedAt       time.Time `reform:"created_at"`
	UpdatedAt       time.Time `reform:"updated_at"`
	ProductTypeName string    `reform:"product_type_name"`
	Concentration1  float64   `reform:"concentration1"`
	Concentration2  float64   `reform:"concentration2"`
	Concentration3  float64   `reform:"concentration3"`
	Note            *string   `reform:"note"`
}
