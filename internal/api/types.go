package api

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
	Place  int
}
