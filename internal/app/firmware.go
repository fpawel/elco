package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/data/chipmem"
	"github.com/fpawel/elco/internal/pkg"
	"github.com/fpawel/elco/internal/pkg/intrng"
	"github.com/hako/durafmt"
	"github.com/powerman/structlog"
	"sort"
	"strings"
	"time"
)

type helperWriteParty struct {
	bytes          map[int][]byte
	failedProducts map[int]error
}

func newHelperWriteParty() helperWriteParty {
	return helperWriteParty{
		bytes:          map[int][]byte{},
		failedProducts: map[int]error{},
	}
}

func writePartyFirmware(x worker) error {

	startTime := time.Now()
	party := data.LastParty()
	products := data.ProductsWithProduction(data.LastPartyID())
	if len(products) == 0 {
		return merry.New("не выбрано ни одного прибора")
	}

	x.log = pkg.LogPrependSuffixKeys(x.log,
		"party_id", party.PartyID,
		"products", formatProducts(products),
	)

	hlp := newHelperWriteParty()

	blockProducts := groupProductsByBlocks(products)
	for _, products := range blockProducts {
		if err := hlp.writeBlock(x, products); err != nil {
			return merry.Wrap(err)
		}
	}

	x.log.Info("Write: ok. Verify.",
		"elapsed", pkg.FormatDuration(time.Since(startTime)),
	)

	startTime = time.Now()
	for _, products := range blockProducts {
		var places []int
		for _, p := range products {
			places = append(places, p.Place)
		}
		sort.Ints(places)
		hlp.verifyProductsFirmware(x, places)
	}

	x.log.Debug("verify complete",
		"elapsed", pkg.FormatDuration(time.Since(startTime)),
	)

	notify.LastPartyChanged(x.log.Info, api.LastParty1())
	return hlp.error()
}

func (hlp helperWriteParty) tryPlace(log *structlog.Logger, place int, f func() error) error {
	err := f()
	if err != nil {
		hlp.failedProducts[place] = err
		log.PrintErr(err, "place", data.FormatPlace(place))
	}
	return err
}

func (hlp helperWriteParty) error() error {
	if len(hlp.failedProducts) == 0 {
		return nil
	}
	type placeErr struct {
		place int
		err   error
	}

	var xs []placeErr
	for n, err := range hlp.failedProducts {
		xs = append(xs, placeErr{n, err})
	}
	sort.Slice(xs, func(i, j int) bool {
		return xs[i].place < xs[j].place
	})

	var errs []string
	for _, n := range xs {
		errs = append(errs, fmt.Sprintf("%s: %s", data.FormatPlace(n.place), n.err))
	}
	return merry.New("Непрошитые места: " + strings.Join(errs, ", "))
}

func (hlp helperWriteParty) verifyProductsFirmware(x worker, places []int) {

	for _, place := range places {
		calc, fPlace := hlp.bytes[place]
		if !fPlace {
			continue
		}
		log := pkg.LogPrependSuffixKeys(x.log,
			"place", data.FormatPlace(place))
		_ = hlp.tryPlace(log, place, func() error {
			b, err := readPlaceFirmware(x, place)
			if err != nil {
				return err
			}
			for _, c := range firmwareAddresses {
				calc := calc[c.addr1 : c.addr2+1]
				read := b[c.addr1 : c.addr2+1]
				if !compareBytes(read, calc) {
					return merry.Errorf(
						"место %s: не совпадают данные по адресам %X...%X",
						data.FormatPlace(place), c.addr1, c.addr2).
						WithValue("расчитано", fmt.Sprintf("% X", read)).
						WithValue("записано", fmt.Sprintf("% X", calc))
				}
			}
			return nil
		})
	}
}

func (hlp *helperWriteParty) writeBlock(x worker, products []*data.Product) error {

	block := products[0].Place / 8

	notify.ReadBlock(x.log.Debug, block)
	defer func() {
		notify.ReadBlock(x.log.Debug, -1)
	}()

	startTime := time.Now()

	var placesMask byte
	for _, p := range products {
		place := byte(p.Place) % 8
		placesMask |= 1 << place
	}

	var placesInBlock []int
	for _, p := range products {
		placesInBlock = append(placesInBlock, p.Place%8)
	}
	sort.Ints(placesInBlock)

	x.log = pkg.LogPrependSuffixKeys(x.log,
		"firmware_write_block", block,
		"places_mask", fmt.Sprintf("%08b", placesMask),
		"selected_places", intrng.Format(placesInBlock),
	)

	for _, p := range products {
		prodInfo := new(data.ProductInfo)
		if err := data.DB.FindByPrimaryKeyTo(prodInfo, p.ProductID); err != nil {
			return err
		}
		log := pkg.LogPrependSuffixKeys(x.log, "product_id", p.ProductID, "serial", p.Serial)
		_ = hlp.tryPlace(log, p.Place, func() error {
			firmware, err := chipmem.ProductInfo{*prodInfo}.CalculateFirmware()
			if err == nil {
				hlp.bytes[p.Place] = firmware.Bytes()
			}
			return err
		})
	}

	defer func() {
		x.log.Debug("end write block", "elapsed", pkg.FormatDuration(time.Since(startTime)))
	}()

	for i, c := range firmwareAddresses {
		for _, p := range products {
			if _, f := hlp.failedProducts[p.Place]; f {
				continue
			}
			addr1 := c.addr1
			addr2 := c.addr2
			b := hlp.bytes[p.Place]
			placeInBlock := p.Place % 8
			x := x.withLogKeys("bytes_count", addr2+1-addr1,
				"range", fmt.Sprintf("%X...%X", addr1, addr2),
				"place", data.FormatPlace(p.Place),
				"product_id", p.ProductID,
				"serial", p.Serial)
			if err := sendDataToWrite42(x, block, placeInBlock, b[addr1:addr2+1]); err != nil {
				return merry.Wrap(err)
			}
		}
		if err := writePreparedDataToFlash(x, block, placesMask, c.addr1, int(c.addr2-c.addr1+1)); err != nil {
			return err
		}

		time.Sleep(cfg.Get().WaitFlashStatusDelay)

		if err := waitStatus45(x, block, placesMask); err != nil {
			if e, ok := err.(errorStatus45); ok {
				x.log.PrintErr("не записано",
					"range", fmt.Sprintf("%X...%X", c.addr1, c.addr1),
					"status45", fmt.Sprintf("%02X", e.code),
				)
				hlp.failedProducts[e.place] = merry.Errorf("%X...%X: статус 45: %02X", c.addr1, c.addr1, e.code)
			} else {
				return err
			}
		}

		if i < len(firmwareAddresses)-1 {
			time.Sleep(cfg.Get().ReadRangeDelay)
		}
	}

	for _, p := range products {
		if _, f := hlp.failedProducts[p.Place]; f {
			continue
		}
		p.Firmware = hlp.bytes[p.Place]
		if err := data.DB.Save(p); err != nil {
			return err
		}
	}

	return nil
}

func readPlaceFirmware(x worker, place int) ([]byte, error) {

	notify.ReadPlace(x.log.Debug, place)
	defer func() {
		notify.ReadPlace(x.log.Debug, -1)
	}()

	startTime := time.Now()
	block := place / 8
	placeInBlock := place % 8
	b := make([]byte, chipmem.FirmwareSize)
	for i := range b {
		b[i] = 0xff
	}

	x.log = pkg.LogPrependSuffixKeys(x.log,
		"place", data.FormatPlace(place),
		"chip", cfg.Get().ChipType,
		"total_read_bytes_count", len(b),
	)

	defer func() {
		x.log.Debug("end read firmware",
			"bytes_read", len(b),
			"elapsed", pkg.FormatDuration(time.Since(startTime)))
	}()

	for i, c := range firmwareAddresses {
		count := c.addr2 - c.addr1 + 1
		req := modbus.Request{
			Addr:     modbus.Addr(block) + 101,
			ProtoCmd: 0x44,
			Data: []byte{
				byte(placeInBlock + 1),
				cfg.Get().ChipType.Code(),
				byte(c.addr1 >> 8),
				byte(c.addr1),
				byte(count >> 8),
				byte(count),
			},
		}

		log := pkg.LogPrependSuffixKeys(x.log,
			"range", fmt.Sprintf("%X...%X", c.addr1, c.addr2),
			"bytes_count", count,
		)

		resp, err := req.GetResponse(log, x.Reader2(), func(request, response []byte) (string, error) {
			if len(response) != 10+int(count) {
				return "", comm.Err.Here().Appendf("ожидалось %d байт ответа, получено %d",
					10+int(count), len(response))
			}
			return "", nil
		})
		if err != nil {
			return nil, merry.Wrap(err)
		}

		copy(b[c.addr1:c.addr1+count], resp[8:8+count])
		if i < len(firmwareAddresses)-1 {
			time.Sleep(cfg.Get().ReadRangeDelay)
		}
	}
	return b, nil
}

func writePlaceFirmware(x worker, place int, bytes []byte) error {

	startTime := time.Now()
	block := place / 8
	placeInBlock := place % 8
	placesMask := byte(1) << byte(placeInBlock)

	x.log = pkg.LogPrependSuffixKeys(x.log,
		"place", data.FormatPlace(place),
		"chip", cfg.Get().ChipType,
		"total_write_bytes_count", len(bytes),
	)

	defer func() {
		notify.ReadPlace(x.log.Debug, -1)
		x.log.Debug("end read firmware",
			"elapsed", pkg.FormatDuration(time.Since(startTime)))
	}()

	notify.ReadPlace(x.log.Debug, place)

	doAddresses := func(addr1, addr2 uint16) error {
		x = x.withLogKeys("range", fmt.Sprintf("%X...%X", addr1, addr2),
			"bytes_count", addr2+1-addr1,
		)
		if err := sendDataToWrite42(x, block, placeInBlock, bytes[addr1:addr2+1]); err != nil {
			return merry.Wrap(err)
		}
		if err := writePreparedDataToFlash(x, block, placesMask, addr1, int(addr2-addr1+1)); err != nil {
			return merry.Wrap(err)
		}
		time.Sleep(cfg.Get().WaitFlashStatusDelay)
		if err := waitStatus45(x, block, placesMask); err != nil {
			return merry.Wrap(err)
		}
		return nil
	}
	for _, c := range firmwareAddresses {
		if err := doAddresses(c.addr1, c.addr2); err != nil {
			return merry.Wrap(err)
		}
	}

	h := newHelperWriteParty()
	h.bytes[place] = bytes
	h.verifyProductsFirmware(x, []int{place})
	return h.error()

}

func waitStatus45(x worker, block int, placesMask byte) error {

	notify.ReadBlock(x.log.Debug, block)
	defer func() {
		notify.ReadBlock(x.log.Debug, -1)
	}()

	ctx, _ := context.WithTimeout(x.ctx, cfg.Get().StatusTimeout)
	for {

		select {
		case <-ctx.Done():
			response, err := readStatus45(x, block)
			if err != nil {
				return err
			}
			status := response[2:10]
			for i, b := range status {
				if (1<<byte(i))&placesMask != 0 && b != 0 {
					return comm.Err.Here().Appendf(
						"%s: таймаут %s, статус[%d]=%X",
						data.FormatPlace(block*8+i), durafmt.Parse(cfg.Get().StatusTimeout), i, b)
				}
			}
			return nil

		default:
			response, err := readStatus45(x, block)
			if err != nil {
				return err
			}
			status := response[2:10]
			statusOk := true
			for i, b := range status {
				if (1<<byte(i))&placesMask != 0 {
					if b == 0 {
						continue
					}
					statusOk = false
					if b != 0xB2 {
						return newErrorStatus45(block*8, i, b)
					}
				}
			}
			if statusOk {
				return nil
			}
		}
	}
}

func readStatus45(x worker, block int) ([]byte, error) {

	x.log = pkg.LogPrependSuffixKeys(x.log, "block", block)
	notify.ReadBlock(x.log.Debug, block)
	defer func() {
		notify.ReadBlock(x.log.Debug, -1)
	}()

	request := modbus.Request{
		Addr:     modbus.Addr(block) + 101,
		ProtoCmd: 0x45,
	}
	return request.GetResponse(x.log, x.Reader2(), func(request, response []byte) (string, error) {
		if len(response) != 12 {
			return "", comm.Err.Here().Appendf("ожидалось 12 байт ответа, получено %d", len(response))
		}
		return "", nil
	})
}

func writePreparedDataToFlash(x worker, block int, placesMask byte, addr uint16, count int) error {

	x.log = pkg.LogPrependSuffixKeys(x.log,
		"block", block,
		"mask", fmt.Sprintf("%08b", placesMask),
		"addr", fmt.Sprintf("%X", addr),
		"bytes_count", count)

	notify.ReadBlock(x.log.Debug, block)
	defer func() {
		notify.ReadBlock(x.log.Debug, -1)
	}()

	req := modbus.Request{
		Addr:     modbus.Addr(block) + 101,
		ProtoCmd: 0x43,
		Data: []byte{
			placesMask,
			cfg.Get().ChipType.Code(),
			byte(addr >> 8),
			byte(addr),
			byte(count >> 8),
			byte(count),
		},
	}
	_, err := req.GetResponse(x.log, x.Reader2(), func(request, response []byte) (string, error) {
		if !compareBytes(response, request) {
			return "", merry.Errorf("запрос не равен ответу: блок измерения %d", block)
		}
		return "", nil
	})
	return merry.Appendf(err, "блок измерения %d", block)
}

func sendDataToWrite42(x worker, block, placeInBlock int, b []byte) error {

	x.log = pkg.LogPrependSuffixKeys(x.log,
		"block", block,
		"place_in_block", placeInBlock,
		"bytes_count", len(b),
	)

	notify.ReadPlace(x.log.Debug, block*8+placeInBlock)
	defer func() {
		notify.ReadPlace(x.log.Debug, -1)
	}()

	req := modbus.Request{
		Addr:     modbus.Addr(block) + 101,
		ProtoCmd: 0x42,
		Data: append([]byte{
			byte(placeInBlock + 1),
			byte(len(b) >> 8),
			byte(len(b)),
		}, b...),
	}

	_, err := req.GetResponse(x.log, x.Reader2(), func(request, response []byte) (string, error) {
		if len(response) != 7 {
			return "", merry.Errorf("длина ответа %d не равна 7: блок измерения %d", len(response), block)
		}
		if !compareBytes(response[:5], request[:5]) {
			return "", merry.Errorf("% X != % X: блок измерения %d", response[:5], request[:5], block)
		}
		return "", nil
	})

	return err
}

func compareBytes(x, y []byte) bool {
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

var firmwareAddresses = []struct{ addr1, addr2 uint16 }{
	{0, 512},
	{1024, 1600},
	{1792, 1810},
	{1824, 1831},
}

type errorStatus45 struct {
	place int
	code  byte
}

func newErrorStatus45(block, placeInBlock int, code byte) errorStatus45 {
	return errorStatus45{block*8 + placeInBlock, code}
}

func (x errorStatus45) Error() string {
	return fmt.Sprintf("%s: статус: %X", data.FormatPlace(x.place), x.code)
}
