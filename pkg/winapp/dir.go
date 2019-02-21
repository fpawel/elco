package winapp

import (
	"github.com/ansel1/merry"
	"github.com/lxn/win"
	"os"
	"os/exec"
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
