package cosmosver

import (
	"github.com/bearnetworkchain/core/ignite/pkg/gomodule"
)

const (
	cosmosModulePath = "github.com/cosmos/cosmos-sdk"
)

// Detect 檢測 Cosmos 的主要版本。
func Detect(appPath string) (version Version, err error) {
	parsed, err := gomodule.ParseAt(appPath)
	if err != nil {
		return version, err
	}

	for _, r := range parsed.Require {
		v := r.Mod

		if v.Path == cosmosModulePath {
			if version, err = Parse(v.Version); err != nil {
				return version, err
			}
		}
	}

	return
}
