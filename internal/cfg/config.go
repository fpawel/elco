package cfg

import (
	"encoding/json"
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/gofins/fins"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var Cfg = &Config{
	d: DefaultDevSettings(),
	u: openGuiSettings(),
}

type Config struct {
	mu sync.Mutex
	u  GuiSettings
	d  DevSettings
}

type GuiSettings struct {
	ComportMeasurer        string
	ComportGas             string
	ChipType               ChipType
	AmbientTemperature     float64
	BlowGasMinutes         int
	HoldTemperatureMinutes int
	EndScaleGas2           bool
}

type ChipType string

const (
	Chip16  = "24LC16"
	Chip64  = "24LC64"
	Chip256 = "24LC256"
)

type DevSettings struct {
	FinsNetwork           FinsNetwork `toml:"fins" comment:"параметры протокола связи с теромкамерой"`
	ComportMeasurer       comm.Config `toml:"measurer" comment:"измерительный блок"`
	ComportGas            comm.Config `toml:"gas_block" comment:"газовый блок"`
	StatusTimeoutSeconds  int         `toml:"status_timeout_seconds" comment:"таймаут статуса прошивки, с"`
	ReadRangeDelayMillis  int         `toml:"read_range_delay_millis" comment:"задержка при считывании диапазонов, мс"`
	WaitFlashStatusDelay  int         `toml:"wait_flash_status_delay_ms" comment:"задержка при считывании статуса записи, мс"`
	ReadBlockPauseSeconds int         `toml:"read_block_pause_seconds" comment:"задержка между опросом блоков измерительных, с"`
}

type FinsNetwork struct {
	MaxAttemptsRead int           `toml:"max_attempts_read" comment:"число попыток получения ответа"`
	TimeoutMS       int           `toml:"timeout_ms" comment:"таймаут считывания, мс"`
	PollSec         time.Duration `toml:"poll_sec" comment:"пауза опроса, с"`
	Server          FinsSettings  `toml:"server" comment:"параметры ссервера omron fins"`
	Client          FinsSettings  `toml:"client" comment:"параметры клиента omron fins"`
}

type FinsSettings struct {
	IP      string `toml:"ip" comment:"tcp адрес"`
	Port    int    `toml:"port" comment:"tcp порт"`
	Network byte   `toml:"network" comment:"fins network"`
	Node    byte   `toml:"network" comment:"fins node"`
	Unit    byte   `toml:"network" comment:"fins unit"`
}

func (x DevSettings) WaitFlashStatusDelayMS() time.Duration {
	return time.Duration(x.WaitFlashStatusDelay) * time.Millisecond
}

func (x FinsNetwork) NewFinsClient() (*fins.Client, error) {
	c, err := fins.NewClient(x.Client.Address(), x.Server.Address())
	if err != nil {
		return nil, err
	}
	c.SetTimeoutMs(uint(x.TimeoutMS))
	return c, nil
}

func (x FinsSettings) Address() fins.Address {
	return fins.NewAddress(x.IP, x.Port, x.Network, x.Node, x.Unit)
}

func (x *Config) Dev() DevSettings {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.d
}

func (x *Config) SetDev(c DevSettings) {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.d = c
}

func (x *Config) Gui() GuiSettings {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.u
}

func (x *Config) SetGui(u GuiSettings) {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.u = u
	x.u.save()
}

func DefaultDevSettings() DevSettings {
	return DevSettings{
		WaitFlashStatusDelay:  1000,
		ReadBlockPauseSeconds: 1,
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
			MaxAttemptsRead: 20,
			PollSec:         2,
			TimeoutMS:       1000,
			Server: FinsSettings{
				IP:   "192.168.250.1",
				Port: 9600,
				Node: 1,
			},
			Client: FinsSettings{
				IP:   "192.168.250.3",
				Port: 9600,
				Node: 254,
			},
		},
	}
}

func (x ChipType) Code() byte {
	switch x {
	case Chip16:
		return 0
	case Chip64:
		return 1
	case Chip256:
		return 2
	default:
		panic("bad chip type")
	}
}
func defaultUserConfig() GuiSettings {
	return GuiSettings{
		ChipType:               Chip16,
		ComportMeasurer:        "COM1",
		ComportGas:             "COM2",
		BlowGasMinutes:         5,
		HoldTemperatureMinutes: 120,
	}
}

func configFileName() string {
	return filepath.Join(filepath.Dir(os.Args[0]), "elco.config.json")
}

func openGuiSettings() GuiSettings {
	x := defaultUserConfig()
	b, err := ioutil.ReadFile(configFileName())
	if err == nil {
		err = json.Unmarshal(b, &x)
	}
	if err != nil {
		fmt.Println(
			"не удалось получить настройки приложения из файла конфигурации. Будут применены настройки приложения по умолчанию",
			"error", err, "file", configFileName())
	}
	return x
}

func (x GuiSettings) save() {
	b, err := json.MarshalIndent(&x, "", "    ")
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(configFileName(), b, 0666); err != nil {
		panic(err)
	}
}
