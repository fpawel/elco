package api

import "github.com/fpawel/elco/internal/cfg"

type ConfigSvc struct{}

func (_ *ConfigSvc) UserAppSetts(_ struct{}, r *cfg.UserConfig) error {
	*r = cfg.Cfg.User()
	return nil
}
