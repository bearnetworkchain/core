package checksum

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Sum 從 dirPath 讀取文件，計算每個文件的 sha256 並創建一個新的校驗和
// 在 outPath 中為它們歸檔。
func Sum(dirPath, outPath string) error {
	var b bytes.Buffer

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, info := range files {
		path := filepath.Join(dirPath, info.Name())
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return err
		}

		if _, err := b.WriteString(fmt.Sprintf("%x %s\n", h.Sum(nil), info.Name())); err != nil {
			return err
		}
	}

	return os.WriteFile(outPath, b.Bytes(), 0666)
}

// 二進制返回可執行文件的 SHA256 哈希，文件在 PATH 中按名稱搜索
func Binary(binaryName string) (string, error) {
	// 獲取二進制路徑
	binaryPath, err := exec.LookPath(binaryName)
	if err != nil {
		return "", err
	}
	f, err := os.Open(binaryPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// 字符串連接所有輸入並返回它們的 SHA256 哈希
func Strings(inputs ...string) string {
	h := sha256.New()
	for _, input := range inputs {
		h.Write([]byte(input))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
