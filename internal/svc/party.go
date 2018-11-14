package svc

import "github.com/fpawel/elco/internal/crud/data"

type Party struct {
	Party    data.PartyInfo
	Products []data.ProductInfo
}
