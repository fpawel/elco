package svc

import "github.com/fpawel/elco/internal/crud/data"

type Party struct {
	data.Party
	Products []data.ProductInfo
}
