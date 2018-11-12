package crud

import "sync"

type Repo struct {
	my sync.Mutex
}
