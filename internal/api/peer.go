package api

type PeerSvc struct {
	p PeerNotifier
}

func NewPeerSvc(p PeerNotifier) *PeerSvc {
	return &PeerSvc{p}
}

type PeerNotifier interface {
	OnStarted()
	OnClosed()
}

func (x *PeerSvc) Init(_ struct{}, _ *struct{}) error {
	x.p.OnStarted()
	return nil
}

func (x *PeerSvc) Close(_ struct{}, _ *struct{}) error {
	x.p.OnClosed()
	return nil
}
