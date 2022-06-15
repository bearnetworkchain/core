package cosmosgen

import (
	"path/filepath"
	"strings"

	"github.com/ignite-hq/cli/ignite/pkg/cache"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosanalysis/module"
	"github.com/ignite-hq/cli/ignite/pkg/gomodule"
	"github.com/ignite-hq/cli/ignite/pkg/protopath"
)

const (
	defaultSdkImport     = "github.com/cosmos/cosmos-sdk"
	moduleCacheNamespace = "generate.setup.module"
)

type ModulesInPath struct {
	Path    string
	Modules []module.Module
}

func (g *generator) setup() (err error) {
// Cosmos SDK 託管自己的 x/ 模塊的 proto 文件和一些自己需要的第三方的 proto 文件和
// 區塊鏈應用程序。 Generate 應該知道這些並將它們提供給區塊鏈
// 想要為自己的 proto 生成代碼的應用程序。
//
// 區塊鏈應用可能使用不同版本的 SDK。以下代碼首先確保
// 應用程序的依賴項由“go mod”下載並緩存在本地文件系統下。
// 然後判斷應用使用的是哪個版本的SDK，絕對路徑是什麼
// 其源代碼。
	if err := cmdrunner.
		New(cmdrunner.DefaultWorkdir(g.appPath)).
		Run(g.ctx, step.New(step.Exec("go", "mod", "download"))); err != nil {
		return err
	}

	//解析應用程序的 go.mod 並提取依賴項。
	modfile, err := gomodule.ParseAt(g.appPath)
	if err != nil {
		return err
	}

	g.sdkImport = defaultSdkImport
	//在 mod 文件中查找任何 cosmos-sdk 替換指令
	for _, r := range modfile.Replace {
		if r.Old.Path == defaultSdkImport {
			g.sdkImport = r.New.Path
			break
		}
	}

	g.deps, err = gomodule.ResolveDependencies(modfile)
	if err != nil {
		return err
	}

	// 這是針對用戶應用程序本身的。它可能包含自定義模塊。這是第一個要尋找的地方。
	g.appModules, err = g.discoverModules(g.appPath, g.protoDir)
	if err != nil {
		return err
	}

// 瀏覽用戶應用程序的 Go 依賴項（在 go.mod 中），其中一些可能是託管的
// 用戶區塊鏈可以使用的 Cosmos SDK 模塊。
//
// Cosmos SDK 是所有區塊鏈的依賴項，所以我們肯定會發現所有的模塊
// SDK 在這個過程中也是如此。
//
// 即使依賴項包含一些 SDK 模塊，也不是所有這些模塊都可以被用戶的區塊鏈使用。
// 這很好，我們仍然可以為那些非模塊生成 JS 客戶端，這取決於用戶使用（在 JS 中導入）
// 不使用生成的模塊。
// 未使用的將永遠不會在 JS 環境中得到解決，也不會交付到生產環境中，JS 捆綁器將避免。
//
// TODO(ilgooz): 我們仍然可以實現某種智能過濾來檢測用戶區塊鏈未使用的模塊
// 在某些時候，很高興擁有。
	moduleCache := cache.New[ModulesInPath](g.cacheStorage, moduleCacheNamespace)
	for _, dep := range g.deps {
		cacheKey := cache.Key(dep.Path, dep.Version)
		modulesInPath, err := moduleCache.Get(cacheKey)
		if err != nil && err != cache.ErrorNotFound {
			return err
		}

		if err == cache.ErrorNotFound {
			path, err := gomodule.LocatePath(g.ctx, g.cacheStorage, g.appPath, dep)
			if err != nil {
				return err
			}
			modules, err := g.discoverModules(path, "")
			if err != nil {
				return err
			}

			modulesInPath = ModulesInPath{
				Path:    path,
				Modules: modules,
			}
			if err := moduleCache.Put(cacheKey, modulesInPath); err != nil {
				return err
			}
		}

		g.thirdModules[modulesInPath.Path] = append(g.thirdModules[modulesInPath.Path], modulesInPath.Modules...)
	}

	return nil
}

func (g *generator) resolveInclude(path string) (paths []string, err error) {
	paths = append(paths, filepath.Join(path, g.protoDir))
	for _, p := range g.o.includeDirs {
		paths = append(paths, filepath.Join(path, p))
	}

	includePaths, err := protopath.ResolveDependencyPaths(g.ctx, g.cacheStorage, g.appPath, g.deps,
		protopath.NewModule(g.sdkImport, append([]string{g.protoDir}, g.o.includeDirs...)...))
	if err != nil {
		return nil, err
	}

	paths = append(paths, includePaths...)
	return paths, nil
}

func (g *generator) discoverModules(path, protoDir string) ([]module.Module, error) {
	var filteredModules []module.Module

	modules, err := module.Discover(g.ctx, g.appPath, path, protoDir)
	if err != nil {
		return nil, err
	}

	for _, m := range modules {
		pp := filepath.Join(path, g.protoDir)
		if !strings.HasPrefix(m.Pkg.Path, pp) {
			continue
		}
		filteredModules = append(filteredModules, m)
	}

	return filteredModules, nil
}
