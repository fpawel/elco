package main

import (
	"github.com/fpawel/elco/internal/daemon"
	"github.com/fpawel/goutils/winapp"
	"github.com/hashicorp/go-multierror"
	"github.com/lxn/win"
	"github.com/pkg/errors"
)

func closeAllPeerWindows() (result error) {
	for hWnd := findPeer(); winapp.IsWindow(hWnd); hWnd = findPeer() {
		if win.SendMessage(hWnd, win.WM_CLOSE, 0, 0) != 0 {
			result = multierror.Append(result, errors.New("can not close peer window"))
		}
	}
	return
}

func findPeer() win.HWND {
	return winapp.FindWindow(daemon.PeerWindowClassName)
}
