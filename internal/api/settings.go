package api

import (
	"github.com/fpawel/elco/internal/data"
	"github.com/pelletier/go-toml"
)

type SettingsSvc struct {
	c *data.Config
}

func NewSetsSvc(cfg *data.Config) *SettingsSvc {
	return &SettingsSvc{cfg}
}

func (x *SettingsSvc) Sections(_ struct{}, res *data.ConfigSections) (err error) {
	*res, err = x.c.Sections()
	return nil
}

func (x *SettingsSvc) SetValue(p [3]string, _ *struct{}) error {
	return x.c.SetValue(p[0], p[1], p[2])
}

func (x *SettingsSvc) PredefinedConfig(_ struct{}, r *string) error {
	b, err := toml.Marshal(x.c.Predefined())
	if err != nil {
		return err
	}
	*r = string(b)
	return nil
}

func (x *SettingsSvc) ChangePredefinedConfig(s [1]string, r *string) error {
	var p data.PredefinedConfig
	if err := toml.Unmarshal([]byte(s[0]), &p); err != nil {
		return err
	}
	b, err := toml.Marshal(&p)
	if err != nil {
		return err
	}
	*r = string(b)
	x.c.SetPredefined(p)
	return nil
}

func (x *SettingsSvc) SetDefaultPredefinedConfig(_ struct{}, r *string) error {
	p := data.DefaultPredefinedConfig()
	b, err := toml.Marshal(&p)
	if err != nil {
		return err
	}
	*r = string(b)
	x.c.SetPredefined(p)
	return nil
}

type GetCheckBlocksArg struct {
	Check [12]bool
}

func (x *SettingsSvc) GetCheckBlocks(_ struct{}, r *GetCheckBlocksArg) error {
	r.Check = x.c.User().CheckBlock
	return nil
}

func (x *SettingsSvc) SetCheckBlock(r [2]int, _ *struct{}) error {
	return x.c.SetCheckBlock(r[0], r[1] != 0)
}
