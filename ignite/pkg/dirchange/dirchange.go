package dirchange

import (
	"bytes"
	"crypto/md5"
	"errors"
	"os"
	"path/filepath"

	"github.com/ignite-hq/cli/ignite/pkg/cache"
)

var ErrNoFile = errors.New("指定路徑中沒有文件")

// SaveDirChecksum 將提供的路徑（目錄或文件）的 md5 校驗和保存在提供的緩存中
// 如果 checksumSavePath 目錄不存在，則創建它
// 路徑是相對於 workdir 的，如果 workdir 為空字符串路徑是絕對路徑
func SaveDirChecksum(checksumCache cache.Cache[[]byte], cacheKey string, workdir string, paths ...string) error {
	checksum, err := ChecksumFromPaths(workdir, paths...)
	if err != nil {
		return err
	}

	// 保存校驗和
	return checksumCache.Put(cacheKey, checksum)
}

// HasDirChecksumChanged 計算提供的路徑（目錄或文件）的 md5 校驗和
// 並將其與當前緩存的校驗和進行比較
// 如果校驗和不存在，則返回 true
// 路徑是相對於 workdir 的，如果 workdir 為空字符串路徑是絕對路徑
func HasDirChecksumChanged(checksumCache cache.Cache[[]byte], cacheKey string, workdir string, paths ...string) (bool, error) {
	savedChecksum, err := checksumCache.Get(cacheKey)
	if err == cache.ErrorNotFound {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	// 計算校驗和
	checksum, err := ChecksumFromPaths(workdir, paths...)
	if errors.Is(err, ErrNoFile) {
		// 沒有文件就無法保存校驗和
		// 因此，如果沒有找到文件，這意味著這些文件已被刪除，則目錄已更改
		return true, nil
	} else if err != nil {
		return false, err
	}

	// 比較校驗和
	if bytes.Equal(checksum, savedChecksum) {
		return false, nil
	}

	// 校驗和已更改
	return true, nil
}

// ChecksumFromPaths 根據提供的路徑計算 md5 校驗和
// 路徑是相對於 workdir 的，如果 workdir 為空字符串路徑是絕對路徑
func ChecksumFromPaths(workdir string, paths ...string) ([]byte, error) {
	hash := md5.New()

	// 如果不存在文件，則無法計算哈希
	noFile := true

	// 讀取文件
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			path = filepath.Join(workdir, path)
		}

		// 不存在的路徑被忽略
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return []byte{}, err
		}

		err := filepath.Walk(path, func(subPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 忽略目錄
			if info.IsDir() {
				return nil
			}

			noFile = false

			// 寫入文件內容
			content, err := os.ReadFile(subPath)
			if err != nil {
				return err
			}
			_, err = hash.Write(content)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return []byte{}, err
		}
	}

	if noFile {
		return []byte{}, ErrNoFile
	}

	// 計算校驗和
	return hash.Sum(nil), nil
}
