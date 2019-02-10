package api

import (
	"github.com/fpawel/elco/internal/elco/config"
	"github.com/fpawel/elco/internal/settings"
	"github.com/pelletier/go-toml"
)

type SettingsSvc struct {
	sets *config.Sets
}

func NewSetsSvc(sets *config.Sets) *SettingsSvc {
	return &SettingsSvc{sets}
}

func (x *SettingsSvc) Sections(_ struct{}, res *settings.ConfigSections) error {
	*res = x.sets.Sections()
	return nil
}

func (x *SettingsSvc) SetValue(p [3]string, _ *struct{}) error {
	return x.sets.SetValue(p[0], p[1], p[2])
}

func (x *SettingsSvc) Predefined(_ struct{}, r *string) error {
	b, err := toml.Marshal(x.sets.Config().Predefined)
	if err != nil {
		return err
	}
	*r = string(b)
	return nil
}

func (x *SettingsSvc) ChangePredefined(s [1]string, r *string) error {

	var cfg config.Predefined
	if err := toml.Unmarshal([]byte(s[0]), &cfg); err != nil {
		return err
	}
	x.sets.SetPredefined(cfg)

	b, err := toml.Marshal(x.sets.Config().Predefined)
	if err != nil {
		return err
	}
	*r = string(b)

	return nil
}

func (x *SettingsSvc) SetDefaultPredefined(_ struct{}, r *string) error {
	c := config.PredefinedConfig()
	x.sets.SetPredefined(c)
	b, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	*r = string(b)
	return nil
}
