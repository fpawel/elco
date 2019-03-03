package cfg

import (
	"database/sql"
	"github.com/fpawel/elco/pkg/serial-comm/comm"
	"gopkg.in/reform.v1"
	"sync"
)

type ConfigProperty struct {
	Hint, Name,
	Value,
	Error string
	Min, Max  sql.NullFloat64
	ValueType ValueType
	List      []string
}

type ValueType int

const (
	VtInt ValueType = iota
	VtFloat
	VtString
	VtComportName
	VtBaud
	VtBool
	VtNullFloat
)

type Config struct {
	mu sync.Mutex
	u  *UserConfig
	p  PredefinedConfig
	db *reform.DB
}

type PredefinedConfig struct {
	ComportMeasurer        comm.Config `toml:"measurer" comment:"измерительный блок"`
	ComportGas             comm.Config `toml:"gas_block" comment:"газовый блок"`
	BlowGasMinutes         int         `toml:"blow_gas_minutes" comment:"длительность продувки газа, мин."`
	HoldTemperatureMinutes int         `toml:"hold_temperature_minutes" comment:"длительность выдержки термокамеры, мин."`
	StatusTimeoutSeconds   int         `toml:"status_timeout_seconds" comment:"таймаут статуса прошивки, с"`
	ReadRangeDelayMillis   int         `toml:"read_range_delay_millis" comment:"задержка при считывании диапазонов, мс"`
	VerboseLogging         bool        `toml:"verbose_logging" comment:"подробные сллбщения в консоли"`
}

type ConfigSection struct {
	Name       string
	Hint       string
	Properties []ConfigProperty
}

type ConfigPropertyValue struct {
	Value, Section, Name string
}

type ConfigSections struct {
	Sections []ConfigSection
}

func OpenConfig(db *reform.DB) *Config {
	return &Config{
		p:  defaultPredefinedConfig,
		u:  openUserConfig(),
		db: db,
	}
}

func (x *Config) Predefined() PredefinedConfig {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.p
}

func (x *Config) SetPredefined(c PredefinedConfig) {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.p = c
}

func (x *Config) User() UserConfig {
	x.mu.Lock()
	defer x.mu.Unlock()
	return *x.u
}

func (x *Config) Save() error {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.u.save()
}

func (x *Config) Sections() (ConfigSections, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	r := ConfigSections{}
	c, err := PartyConfigProperties(x.db)
	if err != nil {
		return r, err
	}
	r.Sections = append(x.u.Sections(), ConfigSection{
		Name:       "Party",
		Hint:       "Партия",
		Properties: c,
	})
	return r, nil
}

func (x *Config) SetValue(section, property, value string) error {
	if section == "Party" {
		return SetPartyConfigValue(x.db, property, value)
	}
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.u.setValue(section, property, value)
}

func DefaultPredefinedConfig() PredefinedConfig {
	return defaultPredefinedConfig
}

var defaultPredefinedConfig = PredefinedConfig{
	BlowGasMinutes:         5,
	HoldTemperatureMinutes: 120,
	ComportGas: comm.Config{
		ReadByteTimeoutMillis: 50,
		ReadTimeoutMillis:     1000,
		MaxAttemptsRead:       3,
	},
	ComportMeasurer: comm.Config{
		ReadByteTimeoutMillis: 15,
		ReadTimeoutMillis:     500,
		MaxAttemptsRead:       10,
	},
	StatusTimeoutSeconds: 3,
	ReadRangeDelayMillis: 10,
}
