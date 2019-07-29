package api

import (
	"github.com/fpawel/elco/internal/cfg"
	"github.com/pelletier/go-toml"
)

type ConfigSvc struct{}

func (_ *ConfigSvc) GetGui(_ struct{}, r *cfg.GuiSettings) error {
	*r = cfg.Cfg.Gui()
	return nil
}

func (_ *ConfigSvc) SetGui(r struct{ C cfg.GuiSettings }, _ *struct{}) error {
	cfg.Cfg.SetGui(r.C)
	return nil
}

func (x *ConfigSvc) Dev(_ struct{}, r *string) error {
	b, err := toml.Marshal(cfg.Cfg.Dev())
	if err != nil {
		return err
	}
	*r = string(b)
	return nil
}

func (x *ConfigSvc) SetDev(s [1]string, r *string) error {
	var p cfg.DevSettings
	if err := toml.Unmarshal([]byte(s[0]), &p); err != nil {
		return err
	}
	b, err := toml.Marshal(&p)
	if err != nil {
		return err
	}
	*r = string(b)
	cfg.Cfg.SetDev(p)
	return nil
}

func (x *ConfigSvc) SetDefaultDev(_ struct{}, r *string) error {
	p := cfg.DefaultDevSettings()
	b, err := toml.Marshal(&p)
	if err != nil {
		return err
	}
	*r = string(b)
	cfg.Cfg.SetDev(p)
	return nil
}
