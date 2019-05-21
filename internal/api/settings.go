package api

import (
	"github.com/fpawel/elco/internal/cfg"
	"github.com/pelletier/go-toml"
)

type SettingsSvc struct {
}

func (x *SettingsSvc) Sections(_ struct{}, res *cfg.ConfigSections) (err error) {
	*res, err = cfg.Cfg.Sections()
	return nil
}

func (x *SettingsSvc) SetValue(p [3]string, _ *struct{}) error {
	return cfg.Cfg.SetValue(p[0], p[1], p[2])
}

func (x *SettingsSvc) PredefinedConfig(_ struct{}, r *string) error {
	b, err := toml.Marshal(cfg.Cfg.Predefined())
	if err != nil {
		return err
	}
	*r = string(b)
	return nil
}

func (x *SettingsSvc) ChangePredefinedConfig(s [1]string, r *string) error {
	var p cfg.PredefinedConfig
	if err := toml.Unmarshal([]byte(s[0]), &p); err != nil {
		return err
	}
	b, err := toml.Marshal(&p)
	if err != nil {
		return err
	}
	*r = string(b)
	cfg.Cfg.SetPredefined(p)
	return nil
}

func (x *SettingsSvc) SetDefaultPredefinedConfig(_ struct{}, r *string) error {
	p := cfg.DefaultPredefinedConfig()
	b, err := toml.Marshal(&p)
	if err != nil {
		return err
	}
	*r = string(b)
	cfg.Cfg.SetPredefined(p)
	return nil
}
