package yaml

import (
	"context"
	"errors"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
)

// Marshal 將對象轉換為 YAML 格式的字符串並轉換
// 從路徑到字符串的字節切片字段更具可讀性。
func Marshal(ctx context.Context, obj interface{}, paths ...string) (string, error) {
	requestYaml, err := yaml.MarshalContext(ctx, obj)
	if err != nil {
		return "", err
	}
	file, err := parser.ParseBytes(requestYaml, 0)
	if err != nil {
		return "", err
	}

	// 規範化將字節切片字段轉換為字符串的結構
	for _, path := range paths {
		pathString, err := yaml.PathString(path)
		if err != nil {
			return "", err
		}
		var byteSlice []byte
		err = pathString.Read(strings.NewReader(string(requestYaml)), &byteSlice)
		if err != nil && !errors.Is(err, yaml.ErrNotFoundNode) {
			return "", err
		}
		if err := pathString.ReplaceWithReader(file,
			strings.NewReader(string(byteSlice)),
		); err != nil {
			return "", err
		}
	}

	return file.String(), nil
}
