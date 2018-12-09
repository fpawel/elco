package config

import (
	"encoding/json"
	"github.com/fpawel/bio3/comport"
	"github.com/fpawel/elco/internal/crud"
	"github.com/fpawel/elco/internal/settings"
	"github.com/fpawel/goutils/serial-comm/comm"
	"io/ioutil"
	"log"
	"sync"
)

type Sets struct {
	mu    sync.Mutex
	c     Config
	party crud.LastParty
}

func OpenSets(party crud.LastParty) *Sets {
	sets := &Sets{
		party: party,
		c: Config{
			Predefined: PredefinedConfig(),
		},
	}

	if b, err := ioutil.ReadFile(configFileName()); err == nil {
		err = json.Unmarshal(b, &sets.c.UserConfig)
	} else {
		if ports := comport.AvailablePorts(); len(ports) > 0 {
			sets.c.ComportName = ports[0]
		}
		log.Println("open config:", err)
	}
	return sets
}

func (x *Sets) Config() Config {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.c
}

func (x *Sets) SetUserConfig(c UserConfig) {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.c.UserConfig = c
	if err := x.save(); err != nil {
		log.Println("Sets.SetUserConfig:", err)
	}
}

func (x *Sets) save() error {
	b, err := json.MarshalIndent(&x.c.UserConfig, "", "    ")
	if err != nil {
		panic(err)
	}
	return ioutil.WriteFile(configFileName(), b, 0666)
}

func (x *Sets) Sections() settings.ConfigSections {
	c := x.Config().UserConfig
	return settings.ConfigSections{
		Sections: []settings.ConfigSection{
			{
				Name:       "Party",
				Hint:       "Партия",
				Properties: x.party.ConfigProperties(),
			},
			c.Section(),
		},
	}
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
	cfg := x.Config().UserConfig
	err := cfg.setValue(section, property, value)
	if err == nil {
		x.SetUserConfig(cfg)
	}
	return err
}
func (x *Sets) SetPredefined(predefined Predefined) {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.c.Predefined = predefined
}

func PredefinedConfig() Predefined {
	return Predefined{
		GasSwitcher: GasSwitcher{
			Addr: 100,
			Comm: comm.Config{
				ReadByteTimeoutMillis: 50,
				ReadTimeoutMillis:     1000,
			},
		},
		Measurer: Measurer{
			Comm: comm.Config{
				ReadByteTimeoutMillis: 50,
				ReadTimeoutMillis:     1000,
			},
		},
	}
}
