package app

import (
	"context"
	"sync"
)

type Hardware struct {
	sync.WaitGroup
	continueFunc,
	cancelFunc,
	skipDelayFunc context.CancelFunc
	ctx context.Context
}
