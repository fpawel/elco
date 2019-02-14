package api

import "github.com/fpawel/elco/internal/data"

type TextMessage struct {
	Text  string
	Level Level
}

type Level int

const (
	Info Level = iota
	Warn
	Error
)

type ReadCurrent struct {
	Values []float64
	Block  int
}

type DelayInfo struct {
	Run         bool
	TimeSeconds int
	What        string
}

type ComportEntry struct {
	Port  string
	Error bool
	Msg   string
}

type Party struct {
	data.Party
	IsLast   bool
	Products []data.ProductInfo
}
