package protopath

import (
	"context"
	"errors"
	"path/filepath"

	"golang.org/x/mod/module"

	"github.com/ignite-hq/cli/ignite/pkg/cache"
	"github.com/ignite-hq/cli/ignite/pkg/gomodule"
	"github.com/ignite-hq/cli/ignite/pkg/xfilepath"
)

var (
	globalInclude = xfilepath.List(
		// 這個應該已經被裸協議執行知道了，但無論如何都要添加它以確保。
		xfilepath.JoinFromHome(xfilepath.Path("local/include")),
		// 這個是放置默認 proto 的建議安裝路徑
		// https://grpc.io/docs/protoc-installation/.
		xfilepath.JoinFromHome(xfilepath.Path(".local/include")),
	)
)

// Module 表示一個託管依賴原型路徑的 go 模塊。
type Module struct {
	importPath string
	include    []string
}

// New Module 創建一個新的 go 模塊表示來查找 protoPaths.
func NewModule(importPath string, protoPaths ...string) Module {
	return Module{
		importPath: importPath,
		include:    protoPaths,
	}
}

// ResolveDependencyPaths 為 go 模塊中給定 r 上的模塊解析依賴原型路徑（包括/-I）。
// r 應該是目標 go app 所需包的列表。它用於解析確切的版本
// 目標應用程序使用的 go 模塊。
// 全局依賴也包含在路徑中。
func ResolveDependencyPaths(ctx context.Context, cacheStorage cache.Storage, src string, versions []module.Version, modules ...Module) (paths []string, err error) {
	globalInclude, err := globalInclude()
	if err != nil {
		return nil, err
	}

	paths = append(paths, globalInclude...)

	var importPaths []string

	for _, module := range modules {
		importPaths = append(importPaths, module.importPath)
	}

	vs := gomodule.FilterVersions(versions, importPaths...)

	if len(vs) != len(modules) {
		return nil, errors.New("go.mod 缺少原型模塊")
	}

	for i, v := range vs {
		path, err := gomodule.LocatePath(ctx, cacheStorage, src, v)
		if err != nil {
			return nil, err
		}

		module := modules[i]
		for _, relpath := range module.include {
			paths = append(paths, filepath.Join(path, relpath))
		}
	}

	return paths, nil
}
