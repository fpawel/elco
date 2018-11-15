package svc

import (
	"github.com/fpawel/elco/internal/crud/data"
)

type Party struct {
	data.PartyInfo
	Products []data.ProductInfo
}
