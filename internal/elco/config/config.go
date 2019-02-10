package config

import (
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/elco/internal/settings"
	"github.com/fpawel/goutils/serial-comm/comm"
	"github.com/fpawel/goutils/serial-comm/comport"
	"github.com/pkg/errors"
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
	Firmware    struct {
		StatusTimeoutSeconds int `toml:"status_timeout_seconds" comment:"таймаут статуса прошивки, с"`
	} `toml:"firmware" comment:"программатор"`
}

type WorkConfig struct {
	BlowGasMinutes         int `toml:"blow_gas_minutes" comment:"длительность продувки газа, мин."`
	HoldTemperatureMinutes int `toml:"hold_temperature_minutes" comment:"длительность выдержки термокамеры, мин."`
}

func (x *UserConfig) Sections() []settings.ConfigSection {

	var chipType string

	switch x.Firmware.ChipType {

	case 0:
		chipType = "24LC16"

	case 1:
		chipType = "24LC64"

	default:
		chipType = "24W256"
	}

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
					Value:     chipType,
					List:      []string{"24LC16", "24LC64", "24W256"},
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

			switch value {

			case "24LC16":
				x.Firmware.ChipType = 0

			case "24LC64":
				x.Firmware.ChipType = 1

			case "24W256":
				x.Firmware.ChipType = 2

			default:
				return errors.Errorf("не верный тип микросхемы: %q", value)

			}
			return nil
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
	return elco.AppName.FileName("config.json")
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
