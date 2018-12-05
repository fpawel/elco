package api

import (
	"github.com/fpawel/elco/internal/app/config"
	"github.com/fpawel/elco/internal/settings"
)

type SettingsSvc struct {
	sets *config.Sets
}

func NewSetsSvc(sets *config.Sets) *SettingsSvc {
	return &SettingsSvc{sets}
}

func (x *SettingsSvc) Get(_ struct{}, res *settings.ConfigSections) error {
	*res = x.sets.UserConfig()
	return nil
}

func (x *SettingsSvc) SetValue(p [3]string, _ *struct{}) error {
	return x.sets.SetValue(p[0], p[1], p[2])
}
