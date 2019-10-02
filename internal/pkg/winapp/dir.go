package winapp

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/pkg"
	"github.com/lxn/win"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
)

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
	if err = pkg.EnsureDir(folderPath); err != nil {
		return "", merry.Wrap(err)
	}
	return folderPath, nil
}
