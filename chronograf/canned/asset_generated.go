// +build assetsplaceholder

package canned

func GeneratedAsset(name string) ([]byte, error) {
	return Asset(name)
}

func GeneratedAssetNames() []string {
	return AssetNames()
}
