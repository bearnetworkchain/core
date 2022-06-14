// Package confile 是加載和覆蓋配置文件的助手。
package confile

import (
	"os"
	"path/filepath"
)

// ConfigFile 代表一個配置文件。
type ConfigFile struct {
	creator EncodingCreator
	path    string
}

// New 啟動一個新的 ConfigFile，使用 creator 作為底層 EncodingCreator 進行編碼和
// 解碼出現或將出現在路徑上的配置文件。
func New(creator EncodingCreator, path string) *ConfigFile {
	return &ConfigFile{
		creator: creator,
		path:    path,
	}
}

// 如果路徑上存在文件，則加載將配置文件的內容加載到 v 中。
// 否則沒有任何內容加載到 v 中並且不返回錯誤。
func (c *ConfigFile) Load(v interface{}) error {
	file, err := os.Open(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()
	return c.creator.Create(file).Decode(v)
}

// Save 通過覆蓋之前的內容將 v 保存到配置文件中，它還會創建
// 如果不存在配置文件。
func (c *ConfigFile) Save(v interface{}) error {
	if err := os.MkdirAll(filepath.Dir(c.path), 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(c.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	return c.creator.Create(file).Encode(v)
}
