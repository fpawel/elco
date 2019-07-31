package data

import (
	"fmt"
)

func FormatPlace(place int) string {
	return fmt.Sprintf("%d.%d", place/8+1, place%8+1)
}

func (s Party) Format() string {
	var str string
	if s.Note.Valid {
		str = ", " + s.Note.String
	}
	return fmt.Sprintf("№%d от %s, %s%s", s.PartyID, s.CreatedAt.Format("02.01.2006"), s.ProductTypeName, str)
}
