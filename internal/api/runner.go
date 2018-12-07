package api

type Runner interface {
	RunReadCurrent()
	StopHardware()
}

type RunnerSvc struct {
	Runner Runner
}

func (x *RunnerSvc) RunReadCurrent(_ struct{}, _ *struct{}) error {
	x.Runner.RunReadCurrent()
	return nil
}

func (x *RunnerSvc) StopHardware(_ struct{}, _ *struct{}) error {
	x.Runner.StopHardware()
	return nil
}
