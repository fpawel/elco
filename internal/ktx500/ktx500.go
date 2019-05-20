package ktx500

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/elco/pkg/copydata"
	"github.com/fpawel/gofins/fins"
	"github.com/powerman/structlog"
	"math"
	"time"
)

var (
	log       = structlog.New()
	GetConfig func() cfg.FinsNetwork
)

func read(f func(*fins.Client) error) (err error) {
	config := GetConfig()

	client, err := config.NewFinsClient()
	if err != nil {
		return merry.Append(err, "установка соединения")
	}
	defer client.Close()

	for attempt := 0; attempt < config.MaxAttemptsRead; attempt++ {
		if err = f(client); err != nil {
			time.Sleep(config.PollSec * time.Second)
			continue
		}
		return nil
	}
	return
}

func write(what string, f func(*fins.Client) error) (err error) {
	config := GetConfig()

	client, err := config.NewFinsClient()
	if err != nil {
		return merry.Append(err, "установка соединения")
	}
	defer client.Close()

	for attempt := 0; attempt < config.MaxAttemptsRead; attempt++ {
		if err = f(client); err != nil {
			time.Sleep(config.PollSec * time.Second)
			continue
		}
		break
	}
	if err != nil {
		err = merry.Append(err, what)
		log.PrintErr(err)
		return err
	}
	log.Info(what)
	return nil
}

func ReadTemperature() (temperature float64, err error) {
	_ = write("запрос температуры", func(c *fins.Client) error {
		temperature, err = readTemperature(c)
		return err
	})
	return
}

func readTemperature(c *fins.Client) (float64, error) {
	temperature, err := finsReadFloat(c, 2)
	if err != nil {
		return 0, merry.Append(err, "запрос температуры")
	}
	return temperature, nil
}

func WriteDestination(value float64) error {
	return write(fmt.Sprintf("запись уставки: %v", value), func(c *fins.Client) error {
		return finsWriteFloat(c, 8, value)
	})
}

func WriteOnOff(value bool) error {
	return write(fmt.Sprintf("включение: %v", value), func(c *fins.Client) error {
		return c.BitTwiddle(fins.MemoryAreaWRBit, 0, 0, value)
	})
}

func WriteCoolOnOff(c *fins.Client, value bool) error {
	return write(fmt.Sprintf("включение компрессора: %v", value), func(c *fins.Client) error {
		return c.BitTwiddle(fins.MemoryAreaWRBit, 0, 10, value)
	})
}

func readInfo(x *api.Ktx500Info) error {
	return read(func(c *fins.Client) error {
		var coolOn, on []bool

		temperature, err := readTemperature(c)
		if err != nil {
			return err
		}

		destination, err := finsReadFloat(c, 8)
		if err != nil {
			return merry.Append(err, "запрос уставки")
		}

		on, err = c.ReadBits(fins.MemoryAreaWRBit, 0, 0, 1)
		if err != nil {
			return merry.Append(err, "запрос состояния термокамеры")
		}

		coolOn, err = c.ReadBits(fins.MemoryAreaWRBit, 0, 10, 1)
		if err != nil {
			return merry.Append(err, "запрос состояния компрессора")
		}

		*x = api.Ktx500Info{
			Temperature: math.Round(temperature*100.) / 100.,
			Destination: destination,
			On:          on[0],
			CoolOn:      coolOn[0],
		}
		return nil
	})
}

func TraceTemperature() {

	var (
		x      api.Ktx500Info
		err    error
		errStr string
	)
	w := copydata.NewNotifyWindow(
		elco.ServerWindowClassName+"TraceTemperature",
		elco.PeerWindowClassName, nil, nil)

	notifyErr := func() {
		if errStr == err.Error() {
			return
		}
		errStr = err.Error()
		notify.Ktx500Error(w, errStr)
		log.PrintErr(merry.Append(err, "КТХ-500"))
	}

	for {
		time.Sleep(time.Second * GetConfig().PollSec)

		var y api.Ktx500Info
		err = readInfo(&y)
		if err != nil {
			notifyErr()
			continue
		}

		if eqNfo(y, x) {
			continue
		}

		x = y
		log.Info(fmt.Sprintf("%v", x.Temperature),
			"вкл", x.On,
			"уставка", x.Destination,
			"компрессор", x.CoolOn)
		notify.Ktx500Info(w, x)
	}
}

func eqNfo(x, y api.Ktx500Info) bool {
	if x == y {
		return true
	}
	a, b := x, y
	a.Temperature, b.Temperature = 0, 0
	return a == b && math.Abs(x.Temperature-y.Temperature) < 0.5
}

func finsReadFloat(finsClient *fins.Client, addr int) (float64, error) {
	xs, err := finsClient.ReadBytes(fins.MemoryAreaDMWord, uint16(addr), 2)
	if err != nil {
		return 0, err
	}
	buf := bytes.NewBuffer([]byte{xs[1], xs[0], xs[3], xs[2]})
	var v float32
	if err := binary.Read(buf, binary.LittleEndian, &v); err != nil {
		return 0, err
	}
	return float64(v), nil
}

func finsWriteFloat(finsClient *fins.Client, addr int, value float64) error {

	buf := bytes.NewBuffer(nil)
	_ = binary.Write(buf, binary.LittleEndian, math.Float32bits(float32(value)))
	b := buf.Bytes()
	words := []uint16{
		binary.LittleEndian.Uint16([]byte{b[0], b[1]}),
		binary.LittleEndian.Uint16([]byte{b[2], b[3]}),
	}
	return finsClient.WriteWords(fins.MemoryAreaDMWord, uint16(addr), words)
}
