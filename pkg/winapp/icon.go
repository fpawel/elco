package winapp

import (
	"github.com/ansel1/merry"
	"github.com/lxn/walk"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

func IconFromBytes(iconBytes []byte) (*walk.Icon, error) {
	tmpIconFile, err := ioutil.TempFile("", "temp_icon_*.ico")
	if err != nil {
		return nil, merry.Errorf("app icon: unable to get temp file: %v", err)
	}

	tmpIconFileName := tmpIconFile.Name()

	// clean up
	defer func() {
		if err := os.Remove(tmpIconFile.Name()); err != nil {
			logrus.WithField("file", tmpIconFileName).Errorln("app icon: unable to clean up temp file")
		}
	}()

	if _, err := tmpIconFile.Write(iconBytes); err != nil {
		_ = tmpIconFile.Close()
		return nil, merry.Errorf("unable to write temp file: %v, %s", err, tmpIconFileName)
	}

	if err := tmpIconFile.Close(); err != nil {
		return nil, merry.Errorf("unable to close temp file: %v, %s", err, tmpIconFileName)
	}

	//We load our icon from a temp file.
	ico, err := walk.NewIconFromFile(tmpIconFile.Name())
	if err != nil {
		return nil, merry.Errorf("app icon: unable to load icon from temp file: %v, %s", err, tmpIconFileName)
	}
	return ico, nil
}
