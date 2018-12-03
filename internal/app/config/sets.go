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
	cfg   C
	mu    sync.Mutex
	party crud.LastParty
}

func OpenSets(party crud.LastParty) (*Sets, error) {
	sets := defaultConfig()
	b, err := ioutil.ReadFile(configFileName())
	if err == nil {
		err = json.Unmarshal(b, &sets)
	}

	return &Sets{cfg: sets, party: party}, err
}

func (x *Sets) Config() C {
	x.mu.Lock()
	defer x.mu.Unlock()
	r := x.cfg
	return r
}

func (x *Sets) SetConfig(cfg C) {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.cfg = cfg
	if err := x.save(); err != nil {
		fmt.Println("Sets.SetConfig:", err)
	}
}

func (x *Sets) UserConfig() settings.ConfigSections {
	cfg := x.Config()

	workSets := []settings.ConfigProperty{
		{
			Hint:      "Показывать посылки COM порта",
			Name:      "DumpComport",
			ValueType: settings.VtBool,
			Value:     strconv.FormatBool(cfg.DumpComport),
		},
	}

	for i := range cfg.BlockSelected {
		workSets = append(workSets, settings.ConfigProperty{
			Hint:      "Блок " + strconv.Itoa(i+1),
			Name:      "Block" + strconv.Itoa(i+1),
			ValueType: settings.VtBool,
			Value:     strconv.FormatBool(cfg.BlockSelected[i]),
		})
	}

	return settings.ConfigSections{
		Sections: []settings.ConfigSection{
			{
				Name:       "Party",
				Hint:       "Партия",
				Properties: x.party.ConfigProperties(),
			},
			settings.Comport("ComportHardware", "СОМ порт стенда", cfg.ComportHardware),
			settings.Comport("ComportGas", "СОМ порт газового блока", cfg.ComportGas),
			settings.Comport("ComportTemperature", "СОМ порт термокамеры", cfg.ComportTemperature),
			{
				Name:       "Work",
				Hint:       "Опрос",
				Properties: workSets,
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
