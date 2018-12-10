package api

type Runner interface {
	RunReadCurrent([12]bool)
	StopHardware()
	Continue()
}

type RunnerSvc struct {
	Runner Runner
}

func (x *RunnerSvc) RunReadCurrent(checkPlaces [12]bool, _ *struct{}) error {
	x.Runner.RunReadCurrent(checkPlaces)
	return nil
}

func (x *RunnerSvc) StopHardware(_ struct{}, _ *struct{}) error {
	x.Runner.StopHardware()
	return nil
}

func (x *RunnerSvc) Continue(_ struct{}, _ *struct{}) error {
	x.Runner.Continue()
	return nil
}
