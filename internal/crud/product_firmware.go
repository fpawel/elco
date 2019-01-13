package crud

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/data"
)

type ProductFirmware struct {
	dbContext
}

func (x ProductFirmware) StoredFirmwareInfo(productID int64) (b data.FirmwareBytes, err error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	var p data.Product
	if err = x.dbr.SelectOneTo(&p, `WHERE product_id = ?`, productID); err != nil {
		return
	}

	if len(p.Firmware) == 0 {
		err = merry.New("ЭХЯ не \"прошита\"")
		return
	}
	if len(p.Firmware) < data.FirmwareSize {
		err = merry.New("не верный формат \"прошивки\"")
		return
	}
	copy(b[:], p.Firmware)
	return
}

func (x ProductFirmware) CalculateFirmwareInfo(productID int64) (data.FirmwareInfo, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	var p data.ProductInfo
	if err := x.dbr.SelectOneTo(&p, `WHERE product_id = ?`, productID); err != nil {
		return data.FirmwareInfo{}, err
	}
	return p.FirmwareInfo(), nil
}
