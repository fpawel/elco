package settings

import "database/sql"

type ConfigProperty struct {
	Hint, Name,
	Value,
	DefaultValue,
	Error string
	Min, Max  sql.NullFloat64
	ValueType ValueType
	List      []string
}

type ValueType int

const (
	VtInt ValueType = iota
	VtFloat
	VtString
	VtComportName
	VtBaud
	VtBool
)

type ConfigSection struct {
	Name       string
	Hint       string
	Properties []ConfigProperty
}

type ConfigPropertyValue struct {
	Value, Section, Name string
}

type ConfigSections struct {
	Sections []ConfigSection
}
