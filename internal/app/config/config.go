package config

import (
	"github.com/fpawel/bio3/comport"
	"github.com/fpawel/elco/internal/app"
	"github.com/fpawel/elco/internal/settings"
	"github.com/fpawel/goutils/serial-comm/comm"
	"github.com/fpawel/goutils/serial-comm/modbus"
	"github.com/pkg/errors"
)

type Config struct {
	UserConfig
	Predefined
}

type UserConfig struct {
	ComportName string
}

type Predefined struct {
	Measurer    Measurer    `toml:"measurer" comment:"измерительный блок"`
	GasSwitcher GasSwitcher `toml:"gas_block" comment:"газовый блок"`
}

type Measurer struct {
	Comm comm.Config `toml:"comm" comment:"транспорт"`
}

type GasSwitcher struct {
	Comm comm.Config `toml:"comm" comment:"транспорт"`
	Addr modbus.Addr `toml:"addr" comment:"адрес в информационной сети"`
}

func (x *UserConfig) Section() settings.ConfigSection {

	return settings.ConfigSection{
		Name: "Hardware",
		Hint: "Оборудование",
		Properties: []settings.ConfigProperty{
			{
				Hint:      "COM порт стенда",
				Name:      "ComportName",
				ValueType: settings.VtComportName,
				Value:     x.ComportName,
			},
		},
	}
}

func (x *UserConfig) setValue(section, property, value string) error {

	switch section {
	case "Hardware":
		switch property {
		case "ComportName":
			ports := comport.AvailablePorts()
			if len(ports) == 0 {
				return errors.Errorf("%q: нет доступных СОМ портов")
			}
			for _, s := range ports {
				if s == value {
					x.ComportName = value
					return nil
				}
			}
			return errors.Errorf("%q : invalid COM port", value)

		}
	}
	return errors.Errorf("%q: %q: invalid section/property", section, property)
}

func configFileName() string {
	return app.AppName.FileName("config.json")
}
