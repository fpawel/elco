package ktx500

import (
	"bytes"
	"encoding/binary"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/elco/pkg/copydata"
	"github.com/fpawel/gofins/fins"
	"time"
)

func TraceTemperature(getFinsNetworkConfigFunc func() cfg.FinsNetwork) {
	var (
		appliedConfig            cfg.FinsNetwork
		finsClient               *fins.Client
		err                      error
		temperature, destination float64
		coolOn, on               []bool
	)
	w := copydata.NewNotifyWindow(
		elco.ServerWindowClassName+"TraceTemperature",
		elco.PeerWindowClassName, nil, nil)

	for {
		time.Sleep(time.Second * getFinsNetworkConfigFunc().PollSec)

		finsNetworkConfig := getFinsNetworkConfigFunc()
		if err != nil || finsNetworkConfig != appliedConfig {
			if finsClient != nil {
				finsClient.Close()
			}

			var err error
			finsClient, err = finsNetworkConfig.NewFinsClient()
			if err != nil {
				notify.Ktx500Error(w, "установка связи: "+err.Error())
				continue
			}
			appliedConfig = finsNetworkConfig
		}

		temperature, err = finsReadFloat(finsClient, 2)
		if err != nil {
			notify.Ktx500Error(w, "температура: "+err.Error())
			continue
		}

		destination, err = finsReadFloat(finsClient, 8)
		if err != nil {
			notify.Ktx500Error(w, "уставка: "+err.Error())
			continue
		}

		on, err = finsClient.ReadBits(fins.MemoryAreaWRBit, 0, 0, 1)
		if err != nil {
			notify.Ktx500Error(w, "статус: "+err.Error())
			continue
		}

		coolOn, err = finsClient.ReadBits(fins.MemoryAreaWRBit, 0, 10, 1)
		if err != nil {
			notify.Ktx500Error(w, "компрессор: "+err.Error())
			continue
		}

		notify.Ktx500Info(w, api.Ktx500Info{
			Temperature: temperature,
			Destination: destination,
			On:          on[0],
			CoolOn:      coolOn[0],
		})

	}
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
