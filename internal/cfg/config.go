package cfg

import (
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/elco/internal/pkg/must"
	"github.com/fpawel/gofins/fins"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	config = func() AppConfig {
		def := AppConfig{
			PublicAppConfig: PublicAppConfig{
				ChipType:               Chip16,
				ComportName:            "COM1",
				ComportGasName:         "COM2",
				ComportName2:           "COM1",
				BlowGasMinutes:         5,
				HoldTemperatureMinutes: 120,
			},

			WaitFlashStatusDelay: time.Second,
			ReadBlockPause:       time.Second,
			ComportGas: comm.Config{
				TimeoutEndResponse: 50 * time.Millisecond,
				TimeoutGetResponse: time.Second,
				MaxAttemptsRead:    3,
			},
			Comport: comm.Config{
				TimeoutEndResponse: 15 * time.Millisecond,
				TimeoutGetResponse: 500 * time.Millisecond,
				MaxAttemptsRead:    10,
			},
			StatusTimeout:  3 * time.Second,
			ReadRangeDelay: 10 * time.Millisecond,
			FinsNetwork: FinsNetwork{
				MaxAttemptsRead: 20,
				Pause:           time.Second * 2,
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
		x := def
		b, err := ioutil.ReadFile(filename())
		if err == nil {
			err = yaml.Unmarshal(b, &x)
		}
		if err != nil {
			fmt.Println(
				"не удалось получить настройки приложения из файла конфигурации. Будут применены настройки приложения по умолчанию",
				"error", err, "file", filename())
			x = def
			data, err := yaml.Marshal(&x)
			must.PanicIf(err)
			must.WriteFile(filename(), data, 0666)
		}
		return x
	}()
	mu sync.Mutex
)

type PublicAppConfig struct {
	ComportName            string   `yaml:"comport_name"`
	ComportGasName         string   `yaml:"comport_gas_name"`
	ComportName2           string   `yaml:"comport_name2"`
	LogComm                bool     `yaml:"log_comm"`
	ChipType               ChipType `yaml:"chip_typ"`
	AmbientTemperature     float64  `yaml:"ambient_temperature"`
	BlowGasMinutes         int      `yaml:"blow_gas_minutes"`
	HoldTemperatureMinutes int      `yaml:"hold_temperature_minutes"`
	EndScaleGas2           bool     `yaml:"end_scale_gas2"`
}

type AppConfig struct {
	PublicAppConfig      `yaml:"public"`
	FinsNetwork          FinsNetwork   `yaml:"fins" comment:"параметры протокола связи с теромкамерой"`
	Comport              comm.Config   `yaml:"comport" comment:"настройки приёмопередачи стенда"`
	ComportGas           comm.Config   `yaml:"gas_block" comment:"настройки приёмопередачи газового блока"`
	StatusTimeout        time.Duration `yaml:"status_timeout" comment:"таймаут статуса прошивки, с"`
	ReadRangeDelay       time.Duration `yaml:"read_range_delay" comment:"задержка при считывании диапазонов, мс"`
	WaitFlashStatusDelay time.Duration `yaml:"wait_flash_status_delay" comment:"задержка при считывании статуса записи, мс"`
	ReadBlockPause       time.Duration `yaml:"read_block_pause" comment:"задержка между опросом блоков измерительных, с"`
}

type FinsNetwork struct {
	MaxAttemptsRead int           `yaml:"max_attempts_read" comment:"число попыток получения ответа"`
	TimeoutMS       uint          `yaml:"timeout_ms" comment:"таймаут считывания, мс"`
	Pause           time.Duration `yaml:"pause" comment:"пауза опроса, с"`
	Server          FinsSettings  `yaml:"server" comment:"параметры ссервера omron fins"`
	Client          FinsSettings  `yaml:"client" comment:"параметры клиента omron fins"`
}

type FinsSettings struct {
	IP       string `yaml:"ip" comment:"tcp адрес"`
	Port     int    `yaml:"port" comment:"tcp порт"`
	Network  byte   `yaml:"network" comment:"fins network"`
	Node     byte   `yaml:"node" comment:"fins node"`
	FinsUnit byte   `yaml:"unit" comment:"fins unit"`
}

type ChipType string

const (
	Chip16  = "24LC16"
	Chip64  = "24LC64"
	Chip256 = "24LC256"
)

func (x ChipType) Code() byte {
	switch x {
	case Chip16:
		return 0
	case Chip64:
		return 1
	case Chip256:
		return 2
	default:
		return 0
	}
}

func (x FinsNetwork) NewFinsClient() (*fins.Client, error) {
	c, err := fins.NewClient(x.Client.Address(), x.Server.Address())
	if err != nil {
		return nil, err
	}
	c.SetTimeoutMs(x.TimeoutMS)
	return c, nil
}

func (x FinsSettings) Address() fins.Address {
	return fins.NewAddress(x.IP, x.Port, x.Network, x.Node, x.FinsUnit)
}

func Set(v AppConfig) {
	mu.Lock()
	defer mu.Unlock()

	d, err := yaml.Marshal(&v)
	must.PanicIf(err)
	must.PanicIf(yaml.Unmarshal(d, &config))
	comm.SetEnableLog(config.LogComm)
	must.WriteFile(filename(), d, 0666)

	return
}

func Get() (result AppConfig) {
	mu.Lock()
	defer mu.Unlock()

	d, err := yaml.Marshal(&config)
	must.PanicIf(err)
	must.PanicIf(yaml.Unmarshal(d, &result))
	return
}

func filename() string {
	return filepath.Join(filepath.Dir(os.Args[0]), "elco.config.yaml")
}
