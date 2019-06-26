package assets

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/gohelp/winapp"
	"os"
	"path"
	"strings"
)

func Ensure() error {

	dir, err := winapp.ProfileFolderPath(".elco", "assets")
	if err != nil {
		return merry.Wrap(err)
	}

	elements, err := AssetDir("assets")
	if err != nil {
		return merry.Wrap(err)
	}
	return ensure(dir, "assets", elements)
}

func ensure(folderPath, assetsFolderPath string, elements []string) error {

	if err := winapp.EnsuredDirectory(folderPath); err != nil {
		return merry.Wrap(err)
	}

	for _, elem := range elements {
		elemFilePath := strings.Replace(path.Join(folderPath, elem), "/", "\\", -1)
		elemAssetsPath := path.Join(assetsFolderPath, elem)

		if elements, err := AssetDir(elemAssetsPath); err == nil {
			return ensure(elemFilePath, elemAssetsPath, elements)
		}

		file, err := os.Create(elemFilePath)
		if err != nil {
			return merry.Wrap(err)
		}

		bytes, err := Asset(elemAssetsPath)
		if err != nil {
			return merry.Wrap(err)
		}

		if _, err := file.Write(bytes); err != nil {
			return merry.Wrap(err)
		}

		if err := file.Close(); err != nil {
			return merry.Wrap(err)
		}
	}

	return nil
}
