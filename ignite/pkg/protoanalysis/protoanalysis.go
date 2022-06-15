// Package protoanalysis 提供用於分析 proto 文件和包的工具集。
package protoanalysis

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// ErrImportNotFound 找不到 proto 文件導入時返回。
var ErrImportNotFound = errors.New("未找到原始導入")

const protoFilePattern = "*.proto"

type Cache map[string]Packages // proto dir path-proto 包對。

func NewCache() Cache {
	return make(Cache)
}

// Parse 通過使用給定的 glob 模式找到它們來解析 proto 包。
func Parse(ctx context.Context, cache Cache, path string) (Packages, error) {
	if cache != nil {
		if packages, ok := cache[path]; ok {
			return packages, nil
		}
	}

	parsed, err := parse(ctx, path, protoFilePattern)
	if err != nil {
		return nil, err
	}

	var packages Packages

	for _, pp := range parsed {
		packages = append(packages, build(*pp))
	}

	if cache != nil {
		cache[path] = packages
	}

	return packages, nil
}

// ParseFile在 path 解析一個 proto 文件。
func ParseFile(path string) (File, error) {
	packages, err := Parse(context.Background(), nil, path)
	if err != nil {
		return File{}, err
	}
	files := packages.Files()
	if len(files) != 1 {
		return File{}, errors.New("路徑不指向單個文件或找不到")
	}
	return files[0], nil
}

// HasMessages 檢查 path 下的 proto 包是否包含具有給定名稱的消息。
func HasMessages(ctx context.Context, path string, names ...string) error {
	pkgs, err := Parse(ctx, NewCache(), path)
	if err != nil {
		return err
	}

	hasName := func(name string) error {
		for _, pkg := range pkgs {
			for _, msg := range pkg.Messages {
				if msg.Name == name {
					return nil
				}
			}
		}
		return fmt.Errorf("無效的原始消息名稱 %s", name)
	}

	for _, name := range names {
		if err := hasName(name); err != nil {
			return err
		}
	}
	return nil
}

// IsImported 檢查路徑下的 proto 包是否導入依賴項列表。
func IsImported(path string, dependencies ...string) error {
	f, err := ParseFile(path)
	if err != nil {
		return err
	}

	for _, wantDep := range dependencies {
		found := false
		for _, fileDep := range f.Dependencies {
			if wantDep == fileDep {
				found = true
				break
			}
		}
		if !found {
			return errors.Wrap(ErrImportNotFound, fmt.Sprintf(
				"無效的原型依賴 %s 用於文件 %s", wantDep, path),
			)
		}
	}
	return nil
}
