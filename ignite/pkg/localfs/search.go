package localfs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Search 在 fs 中搜索具有給定 glob 模式的文件，方法是確保
// 返回的文件路徑已排序。
func Search(path, pattern string) ([]string, error) {
	files := make([]string, 0)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return files, nil
		}
		return nil, err
	}

	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		base := filepath.Base(path)
		// 跳過隱藏文件夾
		if f.IsDir() && strings.HasPrefix(base, ".") {
			return filepath.SkipDir
		}
		// 避免檢查目錄
		if f.IsDir() {
			return nil
		}
		// 檢查文件名和模式是否匹配
		matched, err := filepath.Match(pattern, base)
		if err != nil {
			return err
		}
		if matched {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}
