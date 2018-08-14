// +build assetsplaceholder

package server

func GeneratedAsset(name string) ([]byte, error) {
	return Asset(name)
}
