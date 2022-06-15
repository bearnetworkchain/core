package jsondoc

import (
	"encoding/json"

	"github.com/goccy/go-yaml"
)

// Doc 代表一個 JSON 編碼的數據。
type Doc []byte

// ToDocs 將 JSON 編碼數據列表轉換為文檔。
func ToDocs(data [][]byte) []Doc {
	var docs []Doc
	for _, d := range data {
		docs = append(docs, d)
	}
	return docs
}

// MarshalYAML 在 YAML 編組期間將 Doc 轉換為 YAML 編碼數據。
func (d Doc) MarshalYAML() ([]byte, error) {
	var out interface{}
	if err := json.Unmarshal(d, &out); err != nil {
		return nil, err
	}
	return yaml.Marshal(out)
}

// Pretty 將 Doc 轉換為人類可讀的字符串。
func (d Doc) Pretty() (string, error) {
	proposalyaml, err := yaml.Marshal(d)
	return string(proposalyaml), err
}
