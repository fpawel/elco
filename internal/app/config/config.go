package config

import (
	"github.com/fpawel/elco/internal/app"
	"github.com/fpawel/goutils/serial/comport"
	"github.com/pkg/errors"
	"regexp"
	"strconv"
	"time"
)

type C struct {
	ComportHardware,
	ComportTemperature,
	ComportGas comport.Config
	DumpComport   bool
	BlockSelected [12]bool
}

func (x *C) setValue(section, property, value string) error {
	var pC *comport.Config
	switch section {
	case "ComportHardware":
		pC = &x.ComportHardware
	case "ComportTemperature":
		pC = &x.ComportTemperature
	case "ComportGas":
		pC = &x.ComportGas
	}
	if pC != nil {
		switch property {

		case "name":
			pC.Serial.Name = value
			return nil

		case "baud":
			n, err := strconv.Atoi(value)
			pC.Serial.Baud = n
			return err

		case "timeout":
			n, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			pC.Uart.ReadTimeout = time.Millisecond * time.Duration(n)

		case "timeout_byte":
			n, err := strconv.Atoi(value)
			pC.Uart.ReadByteTimeout = time.Millisecond * time.Duration(n)
			return err

		case "max_attempts_read":
			n, err := strconv.Atoi(value)
			pC.Uart.MaxAttemptsRead = n
			return err
		}
	} else {
		switch section {
		case "Work":
			switch property {
			case "DumpComport":
				n, err := strconv.ParseBool(value)
				x.DumpComport = n
				return err
			default:
				if n, ok := parseBlock(property); ok {
					var err error
					x.BlockSelected[n], err = strconv.ParseBool(value)
					return err
				}
			}
		}
	}
	return errors.Errorf("%q: %q: %q: wrong section/property")
}

var reBlock = regexp.MustCompile(`Block(\d+)`)

func parseBlock(s string) (int, bool) {
	bs := reBlock.FindAllStringSubmatch(s, -1)
	if bs != nil && len(bs[0]) == 2 {
		if n, err := strconv.ParseInt(bs[0][1], 10, 8); err == nil && n > -1 && n < 12 {
			return int(n), true
		}
	}
	return 0, false
}

func configFileName() string {
	return app.AppName.FileName("config.json")
}

func defaultConfig() C {
	return C{
		ComportHardware:    comport.DefaultConfig(),
		ComportGas:         comport.DefaultConfig(),
		ComportTemperature: comport.DefaultConfig(),
		BlockSelected: [12]bool{
			true, true, true, true, true, true, true, true,
		},
	}
}
