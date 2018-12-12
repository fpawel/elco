package api

type Runner interface {
	RunReadCurrent([12]bool)
	StopHardware()
	SkipDelay()
	RunMainWork([5]bool)
}

type RunnerSvc struct {
	Runner Runner
}

func (x *RunnerSvc) RunMainWork(workCheck [5]bool, _ *struct{}) error {
	x.Runner.RunMainWork(workCheck)
	return nil
}

func (x *RunnerSvc) RunReadCurrent(checkPlaces [12]bool, _ *struct{}) error {
	x.Runner.RunReadCurrent(checkPlaces)
	return nil
}

func (x *RunnerSvc) StopHardware(_ struct{}, _ *struct{}) error {
	x.Runner.StopHardware()
	return nil
}

func (x *RunnerSvc) SkipDelay(_ struct{}, _ *struct{}) error {
	x.Runner.SkipDelay()
	return nil
}
