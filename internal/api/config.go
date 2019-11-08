package api

import (
	"github.com/fpawel/elco/internal/cfg"
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

func (_ *ConfigSvc) SetYAML(s [1]string, r *string) error {
	err := cfg.SetYAML(s[0])
	*r = cfg.GetYAML()
	return err
}

func (_ *ConfigSvc) GetYAML(_ struct{}, r *string) error {
	*r = cfg.GetYAML()
	return nil
}
