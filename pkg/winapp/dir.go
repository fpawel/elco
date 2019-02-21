package winapp

import (
	"github.com/ansel1/merry"
	"github.com/lxn/win"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"syscall"
)

func EnsuredDirectory(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) { // создать каталог если его нет
		err = os.MkdirAll(dir, os.ModePerm)
	}
	return err
}

func AppDataFolderPath() (string, error) {
	var dir string
	if dir = os.Getenv("MYAPPDATA"); len(dir) == 0 {
		var buf [win.MAX_PATH]uint16
		if !win.SHGetSpecialFolderPath(0, &buf[0], win.CSIDL_APPDATA, false) {
			return "", merry.New("SHGetSpecialFolderPath failed")
		}
		dir = syscall.UTF16ToString(buf[0:])
	}
	return dir, nil
}

func ShowDirInExporer(dir string) error {
	return exec.Command("Explorer.exe", dir).Start()
}

func ProfileFolderPath(elements ...string) (string, error) {

	usr, err := user.Current()
	if err != nil {
		return "", merry.WithMessage(err, "unable to locate user home catalogue")
	}
	if len(elements) == 0 {
		return usr.HomeDir, nil
	}
	elements = append([]string{usr.HomeDir}, elements...)
	folderPath := filepath.Join(elements...)
	if err = EnsuredDirectory(folderPath); err != nil {
		return "", merry.Wrap(err)
	}
	return folderPath, nil
}

func ProfileFileName(elements ...string) (string, error) {
	if len(elements) < 1 {
		return "", merry.New("file name must be set")
	}
	profileFolderPath, err := ProfileFolderPath(elements[:len(elements)-1]...)
	if err != nil {
		return "", merry.Wrap(err)
	}
	return filepath.Join(profileFolderPath, elements[len(elements)-1]), nil
}

func CurrentDirOrProfileFileName(elements ...string) (string, error) {
	if len(elements) < 1 {
		return "", merry.New("file name must be set")
	}
	baseFileName := elements[len(elements)-1]
	fileName := filepath.Join(filepath.Dir(os.Args[0]), baseFileName)
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		fileName, err = ProfileFileName(elements...)
		if err != nil {
			return "", merry.Wrap(err)
		}
	}
	return fileName, nil
}
