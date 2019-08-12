package data

import (
	"fmt"
)

func FormatPlace(place int) string {
	return fmt.Sprintf("%d.%d", place/8+1, place%8+1)
}
