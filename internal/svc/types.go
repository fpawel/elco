package svc

import "github.com/fpawel/elco/internal/data"

type Party struct {
	data.PartyInfo
	Products []data.ProductInfo
}
