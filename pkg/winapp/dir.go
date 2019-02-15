package winapp

import (
	"github.com/lxn/win"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"syscall"
)

func MustDir(dir string) string {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) { // создать каталог если его нет
		err = os.MkdirAll(dir, os.ModePerm)
	}
	if err != nil {
		panic(err)
	}
	return dir
}

func MustAppDataDir(dirs ...string) string {
	var appDataDir string
	if appDataDir = os.Getenv("MYAPPDATA"); len(appDataDir) == 0 {
		var buf [win.MAX_PATH]uint16
		if !win.SHGetSpecialFolderPath(0, &buf[0], win.CSIDL_APPDATA, false) {
			panic("SHGetSpecialFolderPath failed")
		}
		appDataDir = syscall.UTF16ToString(buf[0:])
	}
	args := []string{appDataDir}
	args = append([]string{appDataDir}, dirs...)
	return MustDir(filepath.Join(args...))
}

func MustAppDir(appName string) string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return MustDir(filepath.Join(usr.HomeDir, "."+appName))
}

func ExeDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return dir
}

func AppDataFileName(elem ...string) string {
	return filepath.Join(MustAppDataDir(elem[:len(elem)-1]...), elem[len(elem)-1])
}

func AppFileName(appName, filename string) string {
	if _, err := os.Stat(filepath.Join(ExeDir(), filename)); !os.IsNotExist(err) {
		return filepath.Join(ExeDir(), filename)
	}
	return filepath.Join(MustAppDir(appName), filename)
}

type AnalitpriborAppName string

func (x AnalitpriborAppName) DataDir() string {
	return MustAppDataDir("Аналитприбор", string(x))
}

func (x AnalitpriborAppName) Dir() string {
	return MustAppDir(string(x))
}

func (x AnalitpriborAppName) DataFileName(fileName string) string {
	return AppDataFileName("Аналитприбор", string(x), fileName)
}

func (x AnalitpriborAppName) FileName(fileName string) string {
	return AppFileName(string(x), fileName)
}

func ShowDirInExporer(dir string) error {
	return exec.Command("Explorer.exe", dir).Start()
}
