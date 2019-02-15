package winapp

import (
	"os"
	"path/filepath"
	"time"
)

func ErrorFileName(exeFileName string) (string, error) {
	exeDir := filepath.Dir(exeFileName)

	if _, err := os.Stat(exeDir); err != nil {
		if os.IsNotExist(err) {
			exeDir = filepath.Dir(os.Args[0])
		} else {
			return "", err
		}
	}
	errorsDir := filepath.Join(exeDir, "errors")

	_, err := os.Stat(errorsDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(errorsDir, os.ModePerm)
	}
	if err != nil {
		return "", err
	}
	t := time.Now()
	return filepath.Join(errorsDir, t.Format("2006_01_02_15_04_05.000")+".log"), nil
}
