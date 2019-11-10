package api

import (
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"gopkg.in/yaml.v3"
)

type ConfigSvc struct{}

func (_ *ConfigSvc) GetConfig(_ struct{}, r *cfg.PublicAppConfig) error {
	*r = cfg.Get().PublicAppConfig
	return nil
}

func (_ *ConfigSvc) SetConfig(r struct{ C cfg.PublicAppConfig }, _ *struct{}) error {
	c := cfg.Get()
	c.PublicAppConfig = r.C
	cfg.Set(c)
	return nil
}

type appConfig struct {
	Main  cfg.AppConfig `yaml:"main"`
	Party Party3        `yaml:"party"`
	Units []data.Units  `yaml:"units"`
	Gases []data.Gas    `yaml:"gases"`
}

func (h *ConfigSvc) SetYAML(s [1]string, r *string) error {

	var c appConfig
	if err := yaml.Unmarshal([]byte(s[0]), &c); err != nil {
		return err
	}
	cfg.Set(c.Main)

	dataParty := data.LastParty()
	c.Party.SetupDataParty(&dataParty)
	if err := data.DB.Save(&dataParty); err != nil {
		return err
	}

	for _, gas := range c.Gases {
		if err := data.DB.Save(&gas); err != nil {
			return err
		}
	}

	for _, units := range c.Units {
		if err := data.DB.Save(&units); err != nil {
			return err
		}
	}

	return h.GetYAML(struct{}{}, r)
}

func (_ *ConfigSvc) GetYAML(_ struct{}, r *string) error {
	var c appConfig
	c.Main = cfg.Get()
	c.Party = newParty3(data.LastParty())
	c.Units = data.ListUnits()
	c.Gases = data.ListGases()

	b, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}
	*r = string(b)
	return nil
}
