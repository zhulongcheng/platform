// +build assetsplaceholder

package dist

import "os"

func GeneratedAsset(name string) ([]byte, error) {
	return Asset(name)
}

func GeneratedAssetInfo(name string) (os.FileInfo, error) {
	return AssetInfo(name)
}

func GeneratedAssetDir(name string) ([]string, error) {
	return AssetDir(name)
}
