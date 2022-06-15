package localfs

import (
	"io/fs"
	"os"
)

// MkdirAllReset 與 os.MkdirAll 相同，只是它在創建路徑之前將其刪除。
func MkdirAllReset(path string, perm fs.FileMode) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return os.MkdirAll(path, perm)
}
