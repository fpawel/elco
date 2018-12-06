package api

type Runner interface {
	RunReadCurrent()
}

type RunnerSvc struct {
	Runner Runner
}

func (x *RunnerSvc) RunReadCurrent(_ struct{}, _ *struct{}) error {
	x.Runner.RunReadCurrent()
	return nil
}
