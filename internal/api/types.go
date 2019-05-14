package api

type ReadCurrent struct {
	Values []float64
	Block  int
}

type DelayInfo struct {
	Run         bool
	TimeSeconds int
	What        string
}

type GetCheckBlocksArg struct {
	Check []bool
}
