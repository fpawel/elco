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
