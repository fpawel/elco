package crud

import (
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/internal/firmware"
)

type ProductFirmware struct {
	dbContext
}

func (x ProductFirmware) Stored(productID int64) (*firmware.Bytes, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	var p data.Product
	if err := x.dbr.SelectOneTo(&p, `WHERE product_id = ?`, productID); err != nil {
		return nil, err
	}
	return firmware.FromBytes(p.Firmware, data.ListGases(x.dbr), data.ListUnits(x.dbr))
}

func (x ProductFirmware) Calculate(productID int64) (firmware.ProductFirmwareInfo, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	var p data.ProductInfo
	if err := x.dbr.SelectOneTo(&p, `WHERE product_id = ?`, productID); err != nil {
		return firmware.ProductFirmwareInfo{}, err
	}
	return firmware.CalculateProductFirmwareInfo(p), nil
}
