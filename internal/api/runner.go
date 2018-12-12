package api

import "errors"

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

	for _, v := range workCheck {
		if v {
			x.Runner.RunMainWork(workCheck)
			return nil
		}
	}
	return errors.New("необходимо отметить как минимум одину настроечную операцию")

}

func (x *RunnerSvc) RunReadCurrent(checkPlaces [12]bool, _ *struct{}) error {

	for _, v := range checkPlaces {
		if v {
			x.Runner.RunReadCurrent(checkPlaces)
			return nil
		}
	}
	return errors.New("необходимо отметить как минимум один блок измерительный из двенадцати")
}

func (x *RunnerSvc) StopHardware(_ struct{}, _ *struct{}) error {
	x.Runner.StopHardware()
	return nil
}

func (x *RunnerSvc) SkipDelay(_ struct{}, _ *struct{}) error {
	x.Runner.SkipDelay()
	return nil
}
