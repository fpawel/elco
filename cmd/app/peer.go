package main

import (
	"bytes"
	"github.com/fpawel/elco/internal/app"
	"github.com/fpawel/goutils/winapp"
	"github.com/hashicorp/go-multierror"
	"github.com/lxn/win"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
)

const peerWindowClassName = "TElcoMainForm"

func peerFound() bool {
	return winapp.IsWindow(findPeer())
}

func closeAllPeerWindows() (result error) {
	for hWnd := findPeer(); winapp.IsWindow(hWnd); hWnd = findPeer() {
		if win.SendMessage(hWnd, win.WM_CLOSE, 0, 0) != 0 {
			result = multierror.Append(result, errors.New("can not close peer window"))
		}
	}
	return
}

func findPeer() win.HWND {
	return winapp.FindWindow(peerWindowClassName)
}

func runPeer() error {
	const (
		peerAppExe = "elcoui.exe"
	)
	dir := filepath.Dir(os.Args[0])

	if _, err := os.Stat(filepath.Join(dir, peerAppExe)); os.IsNotExist(err) {
		dir = app.AppName.Dir()
	}

	cmd := exec.Command(filepath.Join(dir, peerAppExe))
	cmd.Stdout = os.Stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	return cmd.Start()
}
