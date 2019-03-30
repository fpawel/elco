package data

import (
	"fmt"
)

func (s Product) String2() string {
	return fmt.Sprintf(`place="%d.%d" product_id=%d serial=%d`,
		s.Place/8+1, s.Place%8+1, s.ProductID, s.Serial.Int64)
}

func (s ProductInfo) String2() string {
	return fmt.Sprintf(`place="%d.%d" product_id=%d serial=%d`,
		s.Place/8+1, s.Place%8+1, s.ProductID, s.Serial.Int64)
}

func (s PartyInfo) String2() string {
	return fmt.Sprintf(`party_id=%d`, s.PartyID)
}

func (s Party) String2() string {
	return fmt.Sprintf(`party_id=%d`, s.PartyID)
}

func FormatPlace(place int) string {
	return fmt.Sprintf("%d.%d", place/8+1, place%8+1)
}
