package api

import "errors"

type Runner interface {
	RunReadCurrent()
	StopHardware()
	SkipDelay()
	RunTemperature([3]bool)
	RunWritePartyFirmware()
	RunMainError()
}

type RunnerSvc struct {
	Runner Runner
}

func (x *RunnerSvc) RunWritePartyFirmware(_ struct{}, _ *struct{}) error {
	x.Runner.RunWritePartyFirmware()
	return nil
}

func (x *RunnerSvc) RunTemperature(workCheck [3]bool, _ *struct{}) error {
	for _, v := range workCheck {
		if v {
			x.Runner.RunTemperature(workCheck)
			return nil
		}
	}
	return errors.New("необходимо отметить как минимум одну теммпературу")
}

func (x *RunnerSvc) RunReadCurrent(_ struct{}, _ *struct{}) error {

	x.Runner.RunReadCurrent()
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

func (x *RunnerSvc) RunMainError(_ struct{}, _ *struct{}) error {
	x.Runner.RunMainError()
	return nil
}
