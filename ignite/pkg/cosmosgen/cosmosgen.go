package cosmosgen

import (
	"context"
	"path/filepath"

	gomodmodule "golang.org/x/mod/module"

	"github.com/bearnetworkchain/core/ignite/pkg/cache"
	"github.com/bearnetworkchain/core/ignite/pkg/cosmosanalysis/module"
	"github.com/bearnetworkchain/core/ignite/pkg/gomodulepath"
)

// generateOptions 用於配置代碼生成。
type generateOptions struct {
	includeDirs []string
	gomodPath   string

	jsOut               func(module.Module) string
	jsIncludeThirdParty bool
	vuexStoreRootPath   string

	specOut string

	dartOut               func(module.Module) string
	dartIncludeThirdParty bool
	dartRootPath          string
}

// TODO 添加WithInstall。

// ModulePathFunc 定義返回基於 Cosmos SDK 模塊的路徑的函數類型。
type ModulePathFunc func(module.Module) string

// 選項配置代碼生成。
type Option func(*generateOptions)

// WithJSGeneration 添加 JS 代碼生成。為每個模塊調用 out hook
// 檢索應該用於將生成的 js 代碼放入給定模塊的路徑。
// 如果 includeThirdPartyModules 設置為 true，將為第 3 方模塊生成代碼
// 由應用程序（包括 SDK）使用。
func WithJSGeneration(includeThirdPartyModules bool, out ModulePathFunc) Option {
	return func(o *generateOptions) {
		o.jsOut = out
		o.jsIncludeThirdParty = includeThirdPartyModules
	}
}

// WithVuexGeneration 添加了 Vuex 代碼生成。 storeRootPath 用於確定生成的根路徑
// Vuex 商店。 includeThirdPartyModules 和 out 配置底層 JS 庫生成，即
// 記錄在 WithJSGeneration 中。
func WithVuexGeneration(includeThirdPartyModules bool, out ModulePathFunc, storeRootPath string) Option {
	return func(o *generateOptions) {
		o.jsOut = out
		o.jsIncludeThirdParty = includeThirdPartyModules
		o.vuexStoreRootPath = storeRootPath
	}
}

func WithDartGeneration(includeThirdPartyModules bool, out ModulePathFunc, rootPath string) Option {
	return func(o *generateOptions) {
		o.dartOut = out
		o.dartIncludeThirdParty = includeThirdPartyModules
		o.dartRootPath = rootPath
	}
}

// WithGoGeneration添加 Go 代碼生成。
func WithGoGeneration(gomodPath string) Option {
	return func(o *generateOptions) {
		o.gomodPath = gomodPath
	}
}

// WithOpenAPIGeneration 添加 OpenAPI 規範生成。
func WithOpenAPIGeneration(out string) Option {
	return func(o *generateOptions) {
		o.specOut = out
	}
}

// IncludeDirs 配置應用程序 proto 使用的第三方 proto 目錄。
// 相對於項目路徑。
func IncludeDirs(dirs []string) Option {
	return func(o *generateOptions) {
		o.includeDirs = dirs
	}
}

// generator 為 sdk 和 sdk 應用程序生成代碼。
type generator struct {
	ctx          context.Context
	cacheStorage cache.Storage
	appPath      string
	protoDir     string
	o            *generateOptions
	sdkImport    string
	deps         []gomodmodule.Version
	appModules   []module.Module
	thirdModules map[string][]module.Module // 應用程序依賴模塊對。
}

// Generate 從位於 appPath 的 SDK 應用程序的 protoDir 生成代碼，並帶有給定的選項。
// protoDir 必須相對於 projectPath。
func Generate(ctx context.Context, cacheStorage cache.Storage, appPath, protoDir string, options ...Option) error {
	g := &generator{
		ctx:          ctx,
		appPath:      appPath,
		protoDir:     protoDir,
		o:            &generateOptions{},
		thirdModules: make(map[string][]module.Module),
		cacheStorage: cacheStorage,
	}

	for _, apply := range options {
		apply(g.o)
	}

	if err := g.setup(); err != nil {
		return err
	}

	if g.o.gomodPath != "" {
		if err := g.generateGo(); err != nil {
			return err
		}
	}

	// js 生成要求源代碼中存在 Go 類型。因為
	// 在生成的 Go 類型上定義的 sdk.Msg 實現。
	// 所以它需要在 Go 代碼生成之後運行。
	if g.o.jsOut != nil {
		if err := g.generateJS(); err != nil {
			return err
		}
	}

	if g.o.dartOut != nil {
		if err := g.generateDart(); err != nil {
			return err
		}
	}

	if g.o.specOut != "" {
		if err := generateOpenAPISpec(g); err != nil {
			return err
		}
	}

	return nil

}

//VuexStoreModulePath 為 Cosmos SDK 模塊生成 Vuex 存儲模塊路徑。
//根路徑用作生成路徑的前綴。
func VuexStoreModulePath(rootPath string) ModulePathFunc {
	return func(m module.Module) string {
		appModulePath := gomodulepath.ExtractAppPath(m.GoModulePath)
		return filepath.Join(rootPath, appModulePath, m.Pkg.Name, "module")
	}
}
