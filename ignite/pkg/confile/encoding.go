package confile

import (
	"encoding/json"
	"io"

	"github.com/goccy/go-yaml"
	"github.com/pelletier/go-toml"
)

// EncodingCreator 定義了一個構造函數來創建一個編碼創造者
// 一個 io.讀寫器。
type EncodingCreator interface {
	Create(io.ReadWriter) EncodeDecoder
}

// EncodeDecoder 結合了編碼器和解碼器。
type EncodeDecoder interface {
	Encoder
	Decoder
}

// Encoder 應該將 v 編碼為 io.作家 給編碼創造者。
type Encoder interface {
	Encode(v interface{}) error
}

// Decoder 應該解碼來自 io.讀者的 v 給編碼創造者。
type Decoder interface {
	Decode(v interface{}) error
}

// Encoding 實現編碼解碼器
type Encoding struct {
	Encoder
	Decoder
}

// NewEncoding 從 e end d 返回一個新的編碼解碼器實現。
func NewEncoding(e Encoder, d Decoder) EncodeDecoder {
	return &Encoding{
		Encoder: e,
		Decoder: d,
	}
}

// DefaultJSONEncodingCreator 為 JSON 編碼實現編碼創造者。
var DefaultJSONEncodingCreator = &JSONEncodingCreator{}

// DefaultYAMLEncodingCreator 為 YAML 編碼實現編碼創造者。
var DefaultYAMLEncodingCreator = &YAMLEncodingCreator{}

// DefaultTOMLEncodingCreator 為 TOML 編碼實現編碼創造者。
var DefaultTOMLEncodingCreator = &TOMLEncodingCreator{}

// JSONEncodingCreator 為 JSON 編碼實現編碼創造者。
type JSONEncodingCreator struct{}

func (e *JSONEncodingCreator) Create(rw io.ReadWriter) EncodeDecoder {
	return NewEncoding(json.NewEncoder(rw), json.NewDecoder(rw))
}

// YAMLEncodingCreator 為 JSON 編碼實現編碼創造者。
type YAMLEncodingCreator struct{}

func (e *YAMLEncodingCreator) Create(rw io.ReadWriter) EncodeDecoder {
	return NewEncoding(yaml.NewEncoder(rw), yaml.NewDecoder(rw))
}

// TOMLEncodingCreator 為 JSON 編碼實現編碼創造者。
type TOMLEncodingCreator struct{}

func (e *TOMLEncodingCreator) Create(rw io.ReadWriter) EncodeDecoder {
	return NewEncoding(toml.NewEncoder(rw), toml.NewDecoder(rw))
}
