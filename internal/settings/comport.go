package settings

import (
	"database/sql"
	"github.com/fpawel/goutils/serial-comm/comm"
	"strconv"
)

func Comport(name, hint string, c comm.Config) ConfigSection {

	return ConfigSection{
		Name: name,
		Hint: hint,
		Properties: []ConfigProperty{
			{
				Name:      "timeout",
				Hint:      "Таймаут посылки, мс",
				ValueType: VtInt,
				Min:       sql.NullFloat64{Float64: 10, Valid: true},
				Max:       sql.NullFloat64{Float64: 10000, Valid: true},
				Value:     strconv.Itoa(c.ReadTimeoutMillis),
			},
			{
				Name:      "timeout_byte",
				Hint:      "Таймаут байта, мс",
				ValueType: VtInt,
				Min:       sql.NullFloat64{Float64: 10, Valid: true},
				Max:       sql.NullFloat64{Float64: 100, Valid: true},
				Value:     strconv.Itoa(c.ReadByteTimeoutMillis),
			},
			{
				Name:      "max_attempts_read",
				Hint:      "Макс. кол-во попыток",
				ValueType: VtInt,
				Min:       sql.NullFloat64{Valid: true},
				Max:       sql.NullFloat64{Float64: 10, Valid: true},
				Value:     strconv.Itoa(c.MaxAttemptsRead),
			},
		},
	}

}
