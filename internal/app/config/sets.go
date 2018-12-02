package config

import (
	"encoding/json"
	"fmt"
	"github.com/fpawel/elco/internal/crud"
	"github.com/fpawel/elco/internal/settings"
	"io/ioutil"
	"strconv"
	"sync"
)

type Sets struct {
	cfg   Config
	mu    sync.Mutex
	party crud.LastParty
}

func OpenSets() (*Sets, error) {
	sets := defaultConfig()
	b, err := ioutil.ReadFile(configFileName())
	if err == nil {
		err = json.Unmarshal(b, &sets)
	}

	return &Sets{cfg: sets}, err
}

func (x *Sets) Config() Config {
	x.mu.Lock()
	defer x.mu.Unlock()
	r := x.cfg
	return r
}

func (x *Sets) SetConfig(cfg Config) {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.cfg = cfg
	if err := x.save(); err != nil {
		fmt.Println("Sets.SetConfig:", err)
	}
}

func (x *Sets) UserConfig() settings.ConfigSections {
	cfg := x.Config()
	return settings.ConfigSections{
		Sections: []settings.ConfigSection{
			settings.Comport("ComportHardware", "СОМ порт стенда", cfg.ComportHardware),
			settings.Comport("ComportGas", "СОМ порт газового блока", cfg.ComportGas),
			settings.Comport("ComportTemperature", "СОМ порт термокамеры", cfg.ComportTemperature),
			{
				Name: "Work",
				Hint: "Опрос",
				Properties: []settings.ConfigProperty{
					{
						Hint:         "Показывать посылки COM порта",
						Name:         "DumpComport",
						DefaultValue: "false",
						ValueType:    settings.VtBool,
						Value:        strconv.FormatBool(cfg.DumpComport),
					},
				},
			},
			{
				Name:       "Party",
				Hint:       "Партия",
				Properties: x.party.ConfigProperties(),
			},
		},
	}
}

func (x *Sets) save() error {
	b, err := json.MarshalIndent(&x.cfg, "", "    ")
	if err != nil {
		panic(err)
	}
	return ioutil.WriteFile(configFileName(), b, 0666)
}

func (x *Sets) Save() error {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.save()
}

func (x *Sets) SetValue(section, property, value string) error {
	if section == "Party" {
		return x.party.SetConfigValue(property, value)
	}
	cfg := x.Config()
	err := cfg.setValue(section, property, value)
	if err == nil {
		x.SetConfig(cfg)
	}
	return err
}
