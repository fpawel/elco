package cfg

import (
	"encoding/json"
	"fmt"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/gohelp/winapp"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"strconv"
)

type UserConfig struct {
	ComportMeasurer    string
	ComportGas         string
	ChipType           int
	AmbientTemperature float64
}

func (x *UserConfig) Sections() []ConfigSection {

	var chipType string

	switch x.ChipType {

	case 0:
		chipType = "24LC16"

	case 1:
		chipType = "24LC64"

	default:
		chipType = "24W256"
	}

	return []ConfigSection{
		{
			Name: "Comport",
			Hint: "СОМ порт",
			Properties: []ConfigProperty{
				{
					Hint:      "Блоки измерения",
					Name:      "MeasurerComm",
					ValueType: VtComportName,
					Value:     x.ComportMeasurer,
				},
				{
					Hint:      "Газовый блок",
					Name:      "Gas",
					ValueType: VtComportName,
					Value:     x.ComportGas,
				},
			},
		},
		{
			Name: "Hardware",
			Hint: "Оборудование",
			Properties: []ConfigProperty{
				{
					Hint:      "Тип микросхемм",
					Name:      "ChipType",
					ValueType: VtString,
					Value:     chipType,
					List:      []string{"24LC16", "24LC64", "24W256"},
				},
				{
					Hint:      "Температура окружающей среды,\"С",
					Name:      "AmbientTemperature",
					ValueType: VtFloat,
					Value:     fmt.Sprintf("%v", x.AmbientTemperature),
				},
			},
		},
	}
}

func (x *UserConfig) setValue(section, property, value string) error {

	switch section {
	case "Hardware":
		switch property {
		case "AmbientTemperature":
			var err error
			x.AmbientTemperature, err = strconv.ParseFloat(value, 64)
			return err

		case "ChipType":

			switch value {

			case "24LC16":
				x.ChipType = 0

			case "24LC64":
				x.ChipType = 1

			case "24W256":
				x.ChipType = 2

			default:
				return errors.Errorf("не верный тип микросхемы: %q", value)

			}
			return nil

		}
	case "Comport":
		switch property {
		case "MeasurerComm":
			if err := comport.CheckPortNameIsValid(value); err != nil {
				return errors.Errorf("%q: %+v", value, err)
			}
			x.ComportMeasurer = value
			return nil
		case "Gas":
			if err := comport.CheckPortNameIsValid(value); err != nil {
				return errors.Errorf("%q: %+v", value, err)
			}
			x.ComportGas = value
			return nil
		}
	}
	return errors.Errorf("%q: %q: invalid section/property", section, property)
}

func defaultUserConfig() *UserConfig {
	return &UserConfig{
		ChipType:        16,
		ComportMeasurer: "COM1",
		ComportGas:      "COM2",
	}
}

func openUserConfig() *UserConfig {
	configFileName, err := winapp.ProfileFileName(".elco", "config.json")
	if err != nil {
		logrus.Errorln(err, configFileName)
		return defaultUserConfig()
	}
	b, err := ioutil.ReadFile(configFileName)
	x := new(UserConfig)
	if err == nil {
		err = json.Unmarshal(b, x)
	}
	if err != nil {
		logrus.Errorln(err, configFileName)
		return defaultUserConfig()
	}
	return x
}

func (x *UserConfig) save() error {
	b, err := json.MarshalIndent(x, "", "    ")
	if err != nil {
		return err
	}
	configFileName, err := winapp.ProfileFileName(".elco", "config.json")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFileName, b, 0666)
}
