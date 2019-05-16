package daemon

import (
	"bytes"
	"encoding/binary"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/elco"
	"github.com/fpawel/elco/pkg/copydata"
	"github.com/l1va/gofins/fins"
	"time"
)

func traceTemperature(getFinsNetworkConfigFunc func() cfg.FinsNetwork) {
	var (
		appliedFinsNetworkConfig cfg.FinsNetwork
		finsClient               *fins.Client
		w                        = copydata.NewNotifyWindow(elco.ServerWindowClassName+"TraceTemperature", elco.PeerWindowClassName, nil, nil)
	)

	for {
		time.Sleep(time.Second)

		finsNetworkConfig := getFinsNetworkConfigFunc()
		if finsNetworkConfig != appliedFinsNetworkConfig {
			if finsClient != nil {
				finsClient.Close()
			}

			var err error
			finsClient, err = finsNetworkConfig.NewFinsClient()
			if err != nil {
				notify.TraceTemperatureError(w, err.Error())
				continue
			}
			appliedFinsNetworkConfig = finsNetworkConfig
		}

		v, err := finsReadFloat(finsClient, 2)
		if err != nil {
			notify.TraceTemperatureError(w, err.Error())
			continue
		}
		notify.TraceTemperatureInfof(w, "%v\"C", v)

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
