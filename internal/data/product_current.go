package data

import (
	"time"
)

//go:generate reform

// ProductCurrent represents a row in product_current table.
//reform:product_current
type ProductCurrent struct {
	ProductCurrentID int64       `reform:"product_current_id,pk"`
	StoredAt         time.Time   `reform:"stored_at"`
	ProductID        int64       `reform:"product_id"`
	Temperature      Temperature `reform:"temperature"`
	Gas              int         `reform:"gas"`
	CurrentValue     float64     `reform:"current_value"`
	Note             string      `reform:"note"`
}
