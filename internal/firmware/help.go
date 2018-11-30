package firmware

import (
	"database/sql"
	"fmt"
	"github.com/fpawel/goutils"
	"strconv"
)

func formatNullInt64(v sql.NullInt64) string {
	if v.Valid {
		return strconv.FormatInt(v.Int64, 10)
	}
	return ""
}

func formatNullFloat64(v sql.NullFloat64, precision int) string {
	if v.Valid {
		return formatFloat(v.Float64, precision)
	}
	return ""
}

func formatFloat(v float64, precision int) string {
	return strconv.FormatFloat(v, 'g', precision, 64)
}

func formatBCD(b []byte, precision int) string {
	if v, ok := goutils.ParseBCD6(b); ok {
		return formatFloat(v, precision)
	} else {
		return fmt.Sprintf("% X", b)
	}
}
