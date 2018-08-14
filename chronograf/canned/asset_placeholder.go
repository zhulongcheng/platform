// +build !assetsplaceholder

package canned

import "errors"

// The functions defined in this file are placeholders
// until we decide how to get the finalized Chronograf assets in platform.

var errTODO = errors.New("You didn't generate assets for the chronograf/canned folder, using placeholders")

func GeneratedAsset(string) ([]byte, error) {
	return nil, errTODO
}

func GeneratedAssetNames() []string {
	return nil
}
