package api

type WorkResult struct {
	WorkName string
	Tag      WorkResultTag
	Message  string
}

type WorkResultTag int

type ReadCurrent struct {
	Values []float64
	Block  int
}

type DelayInfo struct {
	TotalSeconds, ElapsedSeconds int
	What                         string
}

type GetCheckBlocksArg struct {
	Check []bool
}

type Ktx500Info struct {
	Temperature, Destination float64
	TemperatureOn, CoolOn    bool
}
