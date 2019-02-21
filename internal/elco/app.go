package elco

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/pkg/winapp"
	"os"
	"os/user"
	"path/filepath"
)

const (
	AppName               = "elco"
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

func ProfileFolderPath() (string, error) {

	usr, err := user.Current()
	if err != nil {
		return "", merry.WithMessage(err, "unable to locate user home catalogue")
	}
	profileFolderPath := filepath.Join(usr.HomeDir, ".elco")
	err = winapp.EnsuredDirectory(profileFolderPath)
	if err != nil {
		return "", merry.Wrap(err)
	}
	return profileFolderPath, nil
}

func ProfileFileName(baseFileName string) (string, error) {
	profileFolderPath, err := ProfileFolderPath()
	if err != nil {
		return "", merry.Wrap(err)
	}
	return filepath.Join(profileFolderPath, baseFileName), nil
}

func ConfigFileName() (string, error) {
	return ProfileFileName("config.json")
}

func CurrentDirOrProfileFileName(baseFileName string) (string, error) {
	fileName := filepath.Join(filepath.Dir(os.Args[0]), baseFileName)
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		fileName, err = ProfileFileName(baseFileName)
		if err != nil {
			return "", merry.Wrap(err)
		}
	}
	return fileName, nil
}
