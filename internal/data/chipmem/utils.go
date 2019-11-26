package chipmem

import (
	"database/sql"
	"fmt"
	"github.com/fpawel/comm/modbus"
	"strconv"
	"strings"
)

func formatNullInt64(v sql.NullInt64) string {
	if v.Valid {
		return strconv.FormatInt(v.Int64, 10)
	}
	return ""
}

func formatNullFloat64K(v sql.NullFloat64, k float64, precision int) string {
	if v.Valid {
		return formatFloat(v.Float64*k, precision)
	}
	return ""
}

func formatNullFloat64(v sql.NullFloat64, precision int) string {
	return formatNullFloat64K(v, 1, precision)
}

func formatFloat(v float64, precision int) string {
	return strconv.FormatFloat(v, 'f', precision, 64)
}

func formatBCD(b []byte, precision int) string {
	if v, ok := modbus.ParseBCD6(b); ok {
		return formatFloat(v, precision)
	} else {
		return fmt.Sprintf("% X", b)
	}
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.Replace(s, ",", ".", -1), 64)
}

//func parseFloatPtr(s string) (*float64,error) {
//	s = strings.TrimSpace(s)
//	if len(s) == 0{
//		return nil, nil
//	}
//	s =  strings.Replace(s, ",", ".", -1)
//	v, err := strconv.ParseFloat(s, 64)
//	return &v, err
//}
//
//func parseNullFloat(s string) (sql.NullFloat64, error) {
//
//	s = strings.TrimSpace(s)
//	if len(s) == 0{
//		return sql.NullFloat64{}, nil
//	}
//	s =  strings.Replace(s, ",", ".", -1)
//	v, err := strconv.ParseFloat(s, 64)
//	if err != nil {
//		return sql.NullFloat64{v,true}, nil
//	}
//	return sql.NullFloat64{}, err
//}
