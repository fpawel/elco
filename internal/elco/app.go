package elco

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/pkg/winapp"
	"path/filepath"
)

const (
	PipeName              = `\\.\pipe\elco`
	ServerWindowClassName = "ElcoServerWindow"
	PeerWindowClassName   = "TElcoMainForm"
)

func DataFolderPath() (string, error) {
	appDataFolderPath, err := winapp.AppDataFolderPath()
	if err != nil {
		return "", merry.Wrap(err)
	}
	elcoDataFolderPath := filepath.Join(appDataFolderPath, "elco")
	err = winapp.EnsuredDirectory(elcoDataFolderPath)
	if err != nil {
		return "", merry.Wrap(err)
	}
	return elcoDataFolderPath, nil
}
