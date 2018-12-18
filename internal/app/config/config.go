package config

import (
	"github.com/fpawel/elco/internal/app"
	"github.com/fpawel/elco/internal/settings"
	"github.com/fpawel/goutils/serial-comm/comm"
	"github.com/fpawel/goutils/serial-comm/comport"
	"github.com/pkg/errors"
	"strconv"
)

type Config struct {
	UserConfig
	Predefined
}

type UserConfig struct {
	Comport struct {
		Measurer, GasSwitcher string
	}
	Firmware struct {
		ChipType int
	}
}

type Predefined struct {
	Work        WorkConfig  `toml:"work" comment:"автоматическая настройка"`
	Measurer    comm.Config `toml:"measurer" comment:"измерительный блок"`
	GasSwitcher comm.Config `toml:"gas_block" comment:"газовый блок"`
}

type WorkConfig struct {
	BlowGasMinutes         int `toml:"blow_gas_minutes" comment:"длительность продувки газа, мин."`
	HoldTemperatureMinutes int `toml:"hold_temperature_minutes" comment:"длительность выдержки термокамеры, мин."`
}

func (x *UserConfig) Sections() []settings.ConfigSection {

	return []settings.ConfigSection{
		{
			Name: "Comport",
			Hint: "СОМ порт",
			Properties: []settings.ConfigProperty{
				{
					Hint:      "Блоки измерения",
					Name:      "Measurer",
					ValueType: settings.VtComportName,
					Value:     x.Comport.Measurer,
				},
				{
					Hint:      "Газовый блок",
					Name:      "GasSwitcher",
					ValueType: settings.VtComportName,
					Value:     x.Comport.GasSwitcher,
				},
			},
		},
		{
			Name: "Firmware",
			Hint: "\"Прошивка\"",
			Properties: []settings.ConfigProperty{
				{
					Hint:      "Тип микросхемм",
					Name:      "ChipType",
					ValueType: settings.VtString,
					Value:     strconv.Itoa(x.Firmware.ChipType),
					List:      []string{"16", "64", "256"},
				},
			},
		},
	}
}

func (x *UserConfig) setValue(section, property, value string) error {

	switch section {
	case "Firmware":
		switch property {
		case "ChipType":

			if n, err := strconv.ParseInt(value, 10, 16); err == nil {
				x.Firmware.ChipType = int(n)
				return nil
			} else {
				return errors.Errorf("%q: %+v", value, err)
			}
		}
	case "Comport":
		switch property {
		case "Measurer":
			if err := comport.CheckPortAvailable(value); err != nil {
				return errors.Errorf("%q: %+v", value, err)
			}
			x.Comport.Measurer = value
			return nil
		case "GasSwitcher":
			if err := comport.CheckPortAvailable(value); err != nil {
				return errors.Errorf("%q: %+v", value, err)
			}
			x.Comport.GasSwitcher = value
			return nil

		}
	}
	return errors.Errorf("%q: %q: invalid section/property", section, property)
}

func configFileName() string {
	return app.AppName.FileName("config.json")
}

var predefined = Predefined{
	Work: WorkConfig{
		BlowGasMinutes:         5,
		HoldTemperatureMinutes: 120,
	},
	GasSwitcher: comm.Config{
		ReadByteTimeoutMillis: 50,
		ReadTimeoutMillis:     1000,
	},
	Measurer: comm.Config{
		ReadByteTimeoutMillis: 50,
		ReadTimeoutMillis:     1000,
	},
}