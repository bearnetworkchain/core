// Package goanalysis 提供靜態分析 Go 應用程序的工具集
package goanalysis

import (
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

const (
	mainPackage     = "main"
	goFileExtension = ".go"
)

var (
	// ErrMultipleMainPackagesFound 當發現多個主包而只期望一個時返回。
	ErrMultipleMainPackagesFound = errors.New("multiple main packages found")
)

// DiscoverMain 在 path 下找到主要的 Go 包。
func DiscoverMain(path string) (pkgPaths []string, err error) {
	uniquePaths := make(map[string]struct{})

	err = filepath.Walk(path, func(filePath string, f os.FileInfo, err error) error {
		if f.IsDir() || !strings.HasSuffix(filePath, goFileExtension) {
			return err
		}

		parsed, err := parser.ParseFile(token.NewFileSet(), filePath, nil, parser.PackageClauseOnly)
		if err != nil {
			return err
		}

		if mainPackage == parsed.Name.Name {
			dir := filepath.Dir(filePath)
			uniquePaths[dir] = struct{}{}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	for path := range uniquePaths {
		pkgPaths = append(pkgPaths, path)
	}

	return pkgPaths, nil
}

// DiscoverOneMain 試圖在路徑下只找到一個主要的 Go 包。
func DiscoverOneMain(path string) (pkgPath string, err error) {
	pkgPaths, err := DiscoverMain(path)
	if err != nil {
		return "", err
	}

	count := len(pkgPaths)
	if count == 0 {
		return "", errors.New("找不到主包")
	}
	if count > 1 {
		return "", ErrMultipleMainPackagesFound
	}

	return pkgPaths[0], nil
}

// FindImportedPackages 在 Go 文件中查找導入的包並返回一個映射
// 帶有包名，導入路徑對.
func FindImportedPackages(name string) (map[string]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, nil, 0)
	if err != nil {
		return nil, err
	}

	packages := make(map[string]string) // 名稱 -> 導入
	for _, imp := range f.Imports {
		var importName string
		if imp.Name != nil {
			importName = imp.Name.Name
		} else {
			importParts := strings.Split(imp.Path.Value, "/")
			importName = importParts[len(importParts)-1]
		}

		name := strings.Trim(importName, "\"")
		packages[name] = strings.Trim(imp.Path.Value, "\"")
	}

	return packages, nil
}
