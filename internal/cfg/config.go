package cfg

import (
	"database/sql"
	"github.com/fpawel/comm"
	"github.com/l1va/gofins/fins"
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
	FinsNetwork            FinsNetwork `toml:"fins" comment:"параметры протокола связи с теромкамерой"`
	ComportMeasurer        comm.Config `toml:"measurer" comment:"измерительный блок"`
	ComportGas             comm.Config `toml:"gas_block" comment:"газовый блок"`
	BlowGasMinutes         int         `toml:"blow_gas_minutes" comment:"длительность продувки газа, мин."`
	HoldTemperatureMinutes int         `toml:"hold_temperature_minutes" comment:"длительность выдержки термокамеры, мин."`
	StatusTimeoutSeconds   int         `toml:"status_timeout_seconds" comment:"таймаут статуса прошивки, с"`
	ReadRangeDelayMillis   int         `toml:"read_range_delay_millis" comment:"задержка при считывании диапазонов, мс"`
}

type FinsNetwork struct {
	TimeoutMS int        `toml:"timeout_ms" comment:"таймаут считывания, мс"`
	Server    FinsConfig `toml:"server" comment:"параметры ссервера omron fins"`
	Client    FinsConfig `toml:"client" comment:"параметры клиента omron fins"`
}

type FinsConfig struct {
	IP      string `toml:"ip" comment:"tcp адрес"`
	Port    int    `toml:"port" comment:"tcp порт"`
	Network byte   `toml:"network" comment:"fins network"`
	Node    byte   `toml:"network" comment:"fins node"`
	Unit    byte   `toml:"network" comment:"fins unit"`
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

func (x FinsNetwork) NewFinsClient() (*fins.Client, error) {
	c, err := fins.NewClient(x.Client.Address(), x.Server.Address())
	if err != nil {
		return nil, err
	}
	c.SetTimeoutMs(uint(x.TimeoutMS))
	return c, nil
}

func (x FinsConfig) Address() fins.Address {
	return fins.NewAddress(x.IP, x.Port, x.Network, x.Node, x.Unit)
}

func OpenConfig(db *reform.DB) *Config {
	return &Config{
		p:  DefaultPredefinedConfig(),
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
	return PredefinedConfig{
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
		FinsNetwork: FinsNetwork{
			TimeoutMS: 300,
			Server: FinsConfig{
				IP:   "192.168.250.1",
				Port: 9600,
				Node: 1,
			},
			Client: FinsConfig{
				IP:   "192.168.250.3",
				Port: 9600,
				Node: 254,
			},
		},
	}
}
