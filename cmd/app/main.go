package main

import (
	"flag"
	"github.com/fpawel/elco/internal/daemon"
)

func main() {

	mustRunPeer := true
	flag.BoolVar(&mustRunPeer, "must-run-peer", true, "ensure peer application")
	flag.Parse()

	if mustRunPeer && !peerFound() {
		if err := runPeer(); err != nil {
			panic(err)
		}
	}

	d := daemon.New()
	d.Run(mustRunPeer)

	if err := d.Close(); err != nil {
		panic(err)
	}

	if mustRunPeer {
		if err := closeAllPeerWindows(); err != nil {
			panic(err)
		}
	}
}
