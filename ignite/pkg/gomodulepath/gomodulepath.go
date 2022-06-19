// Package gomodulepath i實現用於操作 Go 模塊路徑的函數。
// 路徑通常定義為域名和包含用戶和路徑的路徑
// 存儲庫名稱，例如“github.com/username/reponame”，但 Go 也允許其他模塊
// 名稱，例如“domain.com/name”、“name”、“namespace/name”或類似的變體。
package gomodulepath

import (
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"

	"github.com/ignite-hq/cli/ignite/pkg/gomodule"
)

//Path 表示 Go 模塊的路徑。
type Path struct {
	// Path 是 Go 模塊的完整路徑。
	// 例如：github.tsum / ignite-hya / tsli.
	RawPath string

	// Root 是 Go 模塊的根目錄名。
	// 例如：github.com/ignite-hq/cli 的 cli。
	Root string

	// Package是可以使用的Go模塊的默認包名
	// 承載模塊的主要功能。
	// 例如：github.com/ignite-hq/cli 的 cli。
	Package string
}

// Parse將 rawpath 解析為模塊 Path。
func Parse(rawpath string) (Path, error) {
	if err := validateRawPath(rawpath); err != nil {
		return Path{}, err
	}
	rootName := root(rawpath)
	// 包名不能包含“-”所以優雅地刪除它們
	// 如果他們在場。
	packageName := stripNonAlphaNumeric(rootName)
	if err := validatePackageName(packageName); err != nil {
		return Path{}, err
	}
	p := Path{
		RawPath: rawpath,
		Root:    rootName,
		Package: packageName,
	}

	return p, nil
}

// ParseAt 解析應用程序的 Go 模塊路徑位於路徑中。
func ParseAt(path string) (Path, error) {
	parsed, err := gomodule.ParseAt(path)
	if err != nil {
		return Path{}, err
	}
	return Parse(parsed.Module.Mod.Path)
}

// Find 在當前路徑和父路徑中搜索 Go 模塊，直到找到它。
func Find(path string) (parsed Path, appPath string, err error) {
	for len(path) != 0 && path != "." && path != "/" {
		parsed, err = ParseAt(path)
		if errors.Is(err, gomodule.ErrGoModNotFound) {
			path = filepath.Dir(path)
			continue
		}
		return parsed, path, err
	}
	return Path{}, "", errors.Wrap(gomodule.ErrGoModNotFound, "找不到您應用的根目錄")
}

// ExtractAppPath 從 Go 模塊路徑中提取應用程序模塊路徑。
func ExtractAppPath(path string) string {
	if path == "" {
		return ""
	}

	items := strings.Split(path, "/")

	// 如果假定為域名，則刪除第一個路徑項
	if len(items) > 1 && strings.Contains(items[0], ".") {
		items = items[1:]
	}

	count := len(items)
	if count == 1 {
		// Go 模塊路徑是單個名稱
		return items[0]
	}

	// 路徑中的最後兩項定義命名空間和應用名稱
	return strings.Join(items[count-2:], "/")
}

func hasDomainNamePrefix(path string) bool {
	if path == "" {
		return false
	}

	// TODO：我們應該使用正則表達式而不是簡單的檢查嗎？
	name, _, _ := strings.Cut(path, "/")
	return strings.Contains(name, ".")
}

func validateRawPath(path string) error {
	// 原始路徑應該是 URI、單個名稱或路徑
	if hasDomainNamePrefix(path) {
		return validateURIPath(path)
	}
	return validateNamePath(path)
}

func validateURIPath(path string) error {
	if err := module.CheckPath(path); err != nil {
		return fmt.Errorf("應用名稱是無效的GO模塊名稱: %w", err)
	}
	return nil
}

func validateNamePath(path string) error {
	if err := module.CheckImportPath(path); err != nil {
		return fmt.Errorf("應用名稱是無效的GO模塊名稱: %w", err)
	}
	return nil
}

func validatePackageName(name string) error {
	fset := token.NewFileSet()
	src := fmt.Sprintf("package %s", name)
	if _, err := parser.ParseFile(fset, "", src, parser.PackageClauseOnly); err != nil {
		// 解析器錯誤在這裡非常低，所以讓我們對用戶完全地隱藏它。

		return errors.New("應用名稱是無效的GO包名稱")
	}
	return nil
}

func root(path string) string {
	sp := strings.Split(path, "/")
	name := sp[len(sp)-1]
	if semver.IsValid(name) { //省略版本。
		name = sp[len(sp)-2]
	}
	return name
}

func stripNonAlphaNumeric(name string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	return strings.ToLower(reg.ReplaceAllString(name, ""))
}
