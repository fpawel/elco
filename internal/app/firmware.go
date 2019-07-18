package app

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/elco/internal/api/notify"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/gohelp"
	"github.com/fpawel/gohelp/helpstr"
	"github.com/fpawel/gohelp/intrng"
	"github.com/hako/durafmt"
	"github.com/powerman/structlog"
	"gopkg.in/reform.v1"
	"sort"
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

func writePartyFirmware() error {

	startTime := time.Now()
	party := data.GetLastParty(data.WithoutProducts)
	products := data.GetLastPartyProducts(data.WithSerials, data.WithProduction)
	if len(products) == 0 {
		return merry.New("не выбрано ни одного прибора")
	}

	log := gohelp.LogWithKeys(log,
		"party_id", party.PartyID,
		"created_at", party.CreatedAt,
		"products", formatProducts(products),
	)

	log.Info("begin")
	defer func() {
		log.Info("end")
	}()

	hlp := newHelperWriteParty()

	blockProducts := GroupProductsByBlocks(products)
	for _, products := range blockProducts {
		if err := hlp.writeBlock(log, products); err != nil {
			return merry.Wrap(err)
		}
	}

	log.Info("Write: ok. Verify.",
		"elapsed", helpstr.FormatDuration(time.Since(startTime)),
	)

	log = gohelp.LogWithKeys(log, "verify", "")

	startTime = time.Now()
	for _, products := range blockProducts {
		var places []int
		for _, p := range products {
			places = append(places, p.Place)
		}
		sort.Ints(places)
		hlp.verifyProductsFirmware(log, places)
	}

	log.Info("verify complete",
		"elapsed", helpstr.FormatDuration(time.Since(startTime)),
	)

	party.Products = data.GetProductsInfoWithPartyID(party.PartyID)
	notify.LastPartyChanged(log, party)

	return hlp.error()
}

func (x helperWriteParty) error() error {
	if len(x.failedProducts) == 0 {
		return nil
	}

	errStr := "Непрошитые места:\n"
	for place, err := range x.failedProducts {
		errStr += "\t- " + data.FormatPlace(place) + ": " + err.Error() + "\n"
	}
	return merry.New(errStr)
}

func (x helperWriteParty) verifyProductsFirmware(log *structlog.Logger, places []int) {
loopPlaces:
	for _, place := range places {
		log := gohelp.LogWithKeys(log, "place", data.FormatPlace(place))
		b, err := readPlaceFirmware(log, place)
		if err != nil {
			x.failedProducts[place] = err
			log.PrintErr(err)
			continue
		}
		for _, c := range firmwareAddresses {
			read := b[c.addr1 : c.addr2+1]
			calc := x.bytes[place][c.addr1 : c.addr2+1]
			if !compareBytes(read, calc) {
				err := merry.Errorf(
					"место %s: не совпадают данные по адресам %X...%X",
					data.FormatPlace(place), c.addr1, c.addr2).
					WithValue("расчитано", fmt.Sprintf("% X", read)).
					WithValue("записано", fmt.Sprintf("% X", calc))
				x.failedProducts[place] = err
				log.PrintErr(err)
				continue loopPlaces
			}
		}
	}
}

func (x *helperWriteParty) writeBlock(log *structlog.Logger, products []*data.Product) error {
	startTime := time.Now()

	block := products[0].Place / 8
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

	log = gohelp.LogWithKeys(log,
		"firmware_write_block", block,
		"places_mask", fmt.Sprintf("%08b", placesMask),
		"selected_places", intrng.Format(placesInBlock),
	)

	log.Info("calculate")

	for _, p := range products {
		prodInfo := new(data.ProductInfo)
		if err := data.DB.FindByPrimaryKeyTo(prodInfo, p.ProductID); err != nil {
			return err
		}
		if firmware, err := prodInfo.Firmware(); err == nil {
			x.bytes[p.Place] = firmware.Bytes()
		} else {
			log.Warn(err, "place", data.FormatPlace(p.Place),
				"product_id", p.ProductID,
				"serial", p.Serial)
			x.failedProducts[p.Place] = err
		}
	}

	log.Info("begin write block")

	defer func() {
		log.Info("end write block", "elapsed", helpstr.FormatDuration(time.Since(startTime)))
	}()

	for i, c := range firmwareAddresses {
		for _, p := range products {
			if _, f := x.failedProducts[p.Place]; f {
				continue
			}
			addr1 := c.addr1
			addr2 := c.addr2
			b := x.bytes[p.Place]
			placeInBlock := p.Place % 8
			log := gohelp.LogWithKeys(log,
				"bytes_count", addr2+1-addr1,
				"range", fmt.Sprintf("%X...%X", addr1, addr2),
				"place", data.FormatPlace(p.Place),
				"product_id", p.ProductID,
				"serial", p.Serial,
			)
			if err := sendDataToWrite42(log, block, placeInBlock, b[addr1:addr2+1]); err != nil {
				return merry.Wrap(err)
			}
		}
		if err := writePreparedDataToFlash(log, block, placesMask, c.addr1, int(c.addr2-c.addr1+1)); err != nil {
			return err
		}

		time.Sleep(cfg.Cfg.Predefined().WaitFlashStatusDelayMS())

		if err := waitStatus45(log, block, placesMask); err != nil {
			return err
		}

		if i < len(firmwareAddresses)-1 {
			time.Sleep(time.Duration(cfg.Cfg.Predefined().ReadRangeDelayMillis) * time.Millisecond)
		}
	}

	for _, p := range products {
		if _, f := x.failedProducts[p.Place]; f {
			continue
		}
		p.Firmware = x.bytes[p.Place]
		if err := data.DB.Save(p); err != nil {
			return err
		}
	}

	return nil
}

func readPlaceFirmware(log *structlog.Logger, place int) ([]byte, error) {

	notify.ReadPlace(log, place)

	startTime := time.Now()
	block := place / 8
	placeInBlock := place % 8
	b := make([]byte, data.FirmwareSize)
	for i := range b {
		b[i] = 0xff
	}

	log = gohelp.LogWithKeys(log,
		"place", data.FormatPlace(place),
		"chip", cfg.Cfg.User().ChipType,
		"total_read_bytes_count", len(b),
	)

	log.Info("begin read firmware")

	defer func() {
		log.Info("end read firmware",
			"bytes_read", len(b),
			"elapsed", helpstr.FormatDuration(time.Since(startTime)))

	}()

	for i, c := range firmwareAddresses {
		count := c.addr2 - c.addr1 + 1
		req := modbus.Request{
			Addr:     modbus.Addr(block) + 101,
			ProtoCmd: 0x44,
			Data: []byte{
				byte(placeInBlock + 1),
				byte(cfg.Cfg.User().ChipType),
				byte(c.addr1 >> 8),
				byte(c.addr1),
				byte(count >> 8),
				byte(count),
			},
		}
		reader := measurerReader(hardware.ctx)

		log := gohelp.LogWithKeys(log,
			"range", fmt.Sprintf("%X...%X", c.addr1, c.addr2),
			"bytes_count", count,
		)

		resp, err := req.GetResponse(log, reader, func(request, response []byte) (string, error) {
			if len(response) != 10+int(count) {
				return "", comm.ErrProtocol.Here().WithMessagef("ожидалось %d байт ответа, получено %d",
					10+int(count), len(response))
			}
			return "", nil
		})
		if err != nil {
			return nil, merry.Wrap(err)
		}

		copy(b[c.addr1:c.addr1+count], resp[8:8+count])
		if i < len(firmwareAddresses)-1 {
			time.Sleep(time.Duration(
				cfg.Cfg.Predefined().ReadRangeDelayMillis) *
				time.Millisecond)
		}
	}

	return b, nil
}

func writePlaceFirmware(log *structlog.Logger, place int, bytes []byte) error {

	notify.ReadPlace(log, place)

	block := place / 8
	placeInBlock := place % 8
	placesMask := byte(1) << byte(placeInBlock)
	startTime := time.Now()

	log = gohelp.LogWithKeys(log,
		"place", data.FormatPlace(place),
		"chip", cfg.Cfg.User().ChipType,
		"total_write_bytes_count", len(bytes),
	)

	log.Info("begin write place firmware")

	defer func() {
		log.Info("end read firmware",
			"elapsed", helpstr.FormatDuration(time.Since(startTime)))
	}()

	doAddresses := func(addr1, addr2 uint16) error {

		log := gohelp.LogWithKeys(log,
			"range", fmt.Sprintf("%X...%X", addr1, addr2),
			"bytes_count", addr2+1-addr1,
		)

		if err := sendDataToWrite42(log, block, placeInBlock, bytes[addr1:addr2+1]); err != nil {
			return merry.Wrap(err)
		}

		if err := writePreparedDataToFlash(log, block, placesMask, addr1, int(addr2-addr1+1)); err != nil {
			return merry.Wrap(err)
		}

		time.Sleep(cfg.Cfg.Predefined().WaitFlashStatusDelayMS())

		if err := waitStatus45(log, block, placesMask); err != nil {
			return merry.Wrap(err)
		}
		return nil
	}

	for _, c := range firmwareAddresses {
		if err := doAddresses(c.addr1, c.addr2); err != nil {
			return merry.Wrap(err)
		}
	}

	var p data.Product
	switch err := data.GetLastPartyProductAtPlace(place, &p); err {
	case nil:
		p.Firmware = bytes
		if err := data.DB.Save(&p); err != nil {
			return err
		}
		log.Info("save")
	case reform.ErrNoRows, sql.ErrNoRows:
		return nil
	default:
		return err
	}

	return nil
}

func waitStatus45(log *structlog.Logger, block int, placesMask byte) error {
	t := time.Duration(cfg.Cfg.Predefined().StatusTimeoutSeconds) * time.Second
	ctx, _ := context.WithTimeout(hardware.ctx, t)
	for {

		select {
		case <-ctx.Done():
			response, err := readStatus45(log, block)
			if err != nil {
				return err
			}
			status := response[2:10]

			for i, b := range status {
				if (1<<byte(i))&placesMask != 0 && b != 0 {
					return comm.ErrProtocol.Here().Appendf(
						"%s: таймаут %s, статус[%d]=%X",
						data.FormatPlace(block*8+i), durafmt.Parse(t), i, b)
				}
			}
			return nil

		default:
			response, err := readStatus45(log, block)
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
						return comm.ErrProtocol.Here().Appendf("место %s: статус[%d]=%X",
							data.FormatPlace(block*8+i), i, b)
					}
				}
			}
			if statusOk {
				return nil
			}
		}
	}
}

func readStatus45(log *structlog.Logger, block int) ([]byte, error) {
	reader := measurerReader(hardware.ctx)
	request := modbus.Request{
		Addr:     modbus.Addr(block) + 101,
		ProtoCmd: 0x45,
	}

	log = gohelp.LogWithKeys(log, "block", block)
	return request.GetResponse(log, reader, func(request, response []byte) (string, error) {
		if len(response) != 12 {
			return "", comm.ErrProtocol.Here().WithMessagef("ожидалось 12 байт ответа, получено %d", len(response))
		}
		return "", nil
	})
}

func writePreparedDataToFlash(log *structlog.Logger, block int, placesMask byte, addr uint16, count int) error {
	req := modbus.Request{
		Addr:     modbus.Addr(block) + 101,
		ProtoCmd: 0x43,
		Data: []byte{
			placesMask,
			byte(cfg.Cfg.User().ChipType),
			byte(addr >> 8),
			byte(addr),
			byte(count >> 8),
			byte(count),
		},
	}
	reader := measurerReader(hardware.ctx)

	log = gohelp.LogWithKeys(log,
		"block", block,
		"mask", fmt.Sprintf("%08b", placesMask),
		"addr", fmt.Sprintf("%X", addr),
		"bytes_count", count)

	_, err := req.GetResponse(log, reader, func(request, response []byte) (string, error) {
		if !compareBytes(response, request) {
			return "", merry.New("запрос не равен ответу")
		}
		return "", nil
	})
	return err
}

func sendDataToWrite42(log *structlog.Logger, block, placeInBlock int, b []byte) error {

	log = gohelp.LogWithKeys(log,
		"block", block,
		"place_in_block", placeInBlock,
		"bytes_count", len(b),
	)

	notify.ReadPlace(log, block*8+placeInBlock)

	req := modbus.Request{
		Addr:     modbus.Addr(block) + 101,
		ProtoCmd: 0x42,
		Data: append([]byte{
			byte(placeInBlock + 1),
			byte(len(b) >> 8),
			byte(len(b)),
		}, b...),
	}

	reader := measurerReader(hardware.ctx)
	_, err := req.GetResponse(log, reader, func(request, response []byte) (string, error) {
		if len(response) != 7 {
			return "", merry.Errorf("длина ответа %d не равна 7", len(response))
		}
		if !compareBytes(response[:5], request[:5]) {
			return "", merry.Errorf("% X != % X", response[:5], request[:5])
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
	{1024, 1535},
	{1536, 1600},
	{1792, 1810},
	{1824, 1831},
}
