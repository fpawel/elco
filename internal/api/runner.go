package api

import "github.com/fpawel/elco/internal/data"

type Runner interface {
	RunReadCurrent()
	StopHardware()
	SkipDelay()
	RunMain(nku, variation, minus, plus bool)
	RunWritePartyFirmware()
	RunSwitchGas(int)
	RunReadAndSaveProductCurrents(field string)
	CopyParty(partyID int64)
}

type RunnerSvc struct {
	Runner Runner
}

func (x *RunnerSvc) RunWritePartyFirmware(_ struct{}, _ *struct{}) error {
	x.Runner.RunWritePartyFirmware()
	return nil
}

func (x *RunnerSvc) RunMain(a [4]bool, _ *struct{}) error {
	x.Runner.RunMain(a[0], a[1], a[2], a[3])
	return nil
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

func (x *RunnerSvc) RunReadAndSaveProductCurrents(r [1]string, _ *struct{}) error {
	x.Runner.RunReadAndSaveProductCurrents(r[0])
	return nil
}

func (x *RunnerSvc) RunSwitchGas(r [1]int, _ *struct{}) error {
	x.Runner.RunSwitchGas(r[0])
	return nil
}

func (x *RunnerSvc) CopyParty(partyID [1]int64, p *data.Party) (err error) {
	go x.Runner.CopyParty(partyID[0])
	return nil
}
