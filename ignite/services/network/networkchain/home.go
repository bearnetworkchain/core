package networkchain

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/bearnetworkchain/core/ignite/services/network/networktypes"
)

// ChainHome 從 SPN 返回用於鏈的默認主目錄。
func ChainHome(launchID uint64) (path string) {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(home, networktypes.SPN, strconv.FormatUint(launchID, 10))
}

// IsChainHomeExist 檢查具有提供的launchID的家是否已經存在。
func IsChainHomeExist(launchID uint64) (path string, ok bool, err error) {
	home := ChainHome(launchID)

	if _, err := os.Stat(home); os.IsNotExist(err) {
		return home, false, nil
	}

	return home, true, nil
}
