package settings

import (
	"database/sql"
	"github.com/fpawel/goutils/serial/comport"
	"strconv"
	"time"
)

func Comport(name, hint string, c comport.Config) ConfigSection {

	return ConfigSection{
		Name: name,
		Hint: hint,
		Properties: []ConfigProperty{
			{
				Name:         "name",
				Hint:         "Имя СОМ порта",
				ValueType:    VtComportName,
				DefaultValue: "COM1",
				Value:        c.Serial.Name,
			},
			{
				Name:         "baud",
				Hint:         "Скорость приёмопередачи, бод",
				ValueType:    VtBaud,
				DefaultValue: "9600",
				Value:        strconv.Itoa(c.Serial.Baud),
			},
			{
				Name:         "timeout",
				Hint:         "Таймаут посылки, мс",
				ValueType:    VtInt,
				Min:          sql.NullFloat64{Float64: 10, Valid: true},
				Max:          sql.NullFloat64{Float64: 10000, Valid: true},
				DefaultValue: "1000",
				Value:        timeMillis(c.Uart.ReadTimeout),
			},
			{
				Name:         "timeout_byte",
				Hint:         "Таймаут байта, мс",
				ValueType:    VtInt,
				Min:          sql.NullFloat64{Float64: 10, Valid: true},
				Max:          sql.NullFloat64{Float64: 100, Valid: true},
				DefaultValue: "50",
				Value:        timeMillis(c.Uart.ReadByteTimeout),
			},
			{
				Name:         "max_attempts_read",
				Hint:         "Макс. кол-во попыток",
				ValueType:    VtInt,
				Min:          sql.NullFloat64{Valid: true},
				Max:          sql.NullFloat64{Float64: 10, Valid: true},
				DefaultValue: "0",
				Value:        strconv.Itoa(c.Uart.MaxAttemptsRead),
			},
		},
	}

}

func timeMillis(t time.Duration) string {
	return strconv.Itoa(int(t / time.Millisecond))
}
