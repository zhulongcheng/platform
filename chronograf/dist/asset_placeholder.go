// +build !assetsplaceholder

package dist

import (
	"errors"
	"os"
)

// The functions defined in this file are placeholders
// until we decide how to get the finalized Chronograf assets in platform.

var errTODO = errors.New("You didn't generate assets for the chronograf/dist folder, using placeholders")

func GeneratedAsset(string) ([]byte, error) {
	return nil, errTODO
}

func GeneratedAssetInfo(name string) (os.FileInfo, error) {
	return nil, errTODO
}

func GeneratedAssetDir(name string) ([]string, error) {
	return nil, errTODO
}

