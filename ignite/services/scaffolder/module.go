package scaffolder

import (
	"context"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/genny"

	"github.com/ignite-hq/cli/ignite/pkg/cache"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"
	appanalysis "github.com/ignite-hq/cli/ignite/pkg/cosmosanalysis/app"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosver"
	"github.com/ignite-hq/cli/ignite/pkg/gocmd"
	"github.com/ignite-hq/cli/ignite/pkg/multiformatname"
	"github.com/ignite-hq/cli/ignite/pkg/placeholder"
	"github.com/ignite-hq/cli/ignite/pkg/validation"
	"github.com/ignite-hq/cli/ignite/pkg/xgenny"
	"github.com/ignite-hq/cli/ignite/templates/field"
	"github.com/ignite-hq/cli/ignite/templates/module"
	modulecreate "github.com/ignite-hq/cli/ignite/templates/module/create"
	moduleimport "github.com/ignite-hq/cli/ignite/templates/module/import"
)

const (
	wasmImport    = "github.com/CosmWasm/wasmd"
	wasmVersion   = "v0.16.0"
	extrasImport  = "github.com/tendermint/spm-extras"
	extrasVersion = "v0.1.0"
	appPkg        = "app"
	moduleDir     = "x"
)

var (
	// reservedNames 是 Cosmos-SDK 應用程序中定義的默認模塊的名稱，或者是默認查詢和 tx CLI 命名空間中使用的名稱
	// 新模塊的名稱不能等於保留名稱
	// 一個映射用於直接比較
	reservedNames = map[string]struct{}{
		"account":      {},
		"auth":         {},
		"authz":        {},
		"bank":         {},
		"block":        {},
		"broadcast":    {},
		"crisis":       {},
		"capability":   {},
		"distribution": {},
		"encode":       {},
		"evidence":     {},
		"feegrant":     {},
		"genutil":      {},
		"gov":          {},
		"group":        {},
		"ibc":          {},
		"mint":         {},
		"multisign":    {},
		"params":       {},
		"sign":         {},
		"slashing":     {},
		"staking":      {},
		"transfer":     {},
		"tx":           {},
		"txs":          {},
		"upgrade":      {},
		"vesting":      {},
	}

	// defaultStoreKeys 是 Cosmos-SDK 應用程序中定義的默認存儲鍵的名稱
	// 由於潛在的存儲鍵衝突，新模塊的名稱不能在其前綴中定義存儲鍵
	defaultStoreKeys = []string{
		"acc",
		"bank",
		"capability",
		"distribution",
		"evidence",
		"feegrant",
		"gov",
		"group",
		"mint",
		"slashing",
		"staking",
		"upgrade",
		"ibc",
		"transfer",
	}
)

// moduleCreationOptions 包含用於創建新模塊的選項
type moduleCreationOptions struct {
	// ibc 如果模塊是 ibc 模塊，則為 true
	ibc bool

	// params 參數列表
	params []string

	// ibcChannelOrdering ibc 通道排序
	ibcChannelOrdering string

	// 模塊依賴的依賴列表
	dependencies []modulecreate.Dependency
}

// ModuleCreationOption 配置鏈。
type ModuleCreationOption func(*moduleCreationOptions)

// WithIBC 搭建一個啟用了 IBC 的模塊
func WithIBC() ModuleCreationOption {
	return func(m *moduleCreationOptions) {
		m.ibc = true
	}
}

// WithParams 用 params 搭建一個模塊
func WithParams(params []string) ModuleCreationOption {
	return func(m *moduleCreationOptions) {
		m.params = params
	}
}

// WithIBCChannelOrdering 配置 IBC 模塊的通道排序
func WithIBCChannelOrdering(ordering string) ModuleCreationOption {
	return func(m *moduleCreationOptions) {
		switch ordering {
		case "ordered":
			m.ibcChannelOrdering = "ORDERED"
		case "unordered":
			m.ibcChannelOrdering = "UNORDERED"
		default:
			m.ibcChannelOrdering = "NONE"
		}
	}
}

// WithDependencies 指定模塊所依賴的模塊名稱
func WithDependencies(dependencies []modulecreate.Dependency) ModuleCreationOption {
	return func(m *moduleCreationOptions) {
		m.dependencies = dependencies
	}
}

// CreateModule 在腳手架應用中創建一個新的空模塊
func (s Scaffolder) CreateModule(
	cacheStorage cache.Storage,
	tracer *placeholder.Tracer,
	moduleName string,
	options ...ModuleCreationOption,
) (sm xgenny.SourceModification, err error) {
	mfName, err := multiformatname.NewName(moduleName, multiformatname.NoNumber)
	if err != nil {
		return sm, err
	}
	moduleName = mfName.LowerCase

	// 檢查模塊名是否有效
	if err := checkModuleName(s.path, moduleName); err != nil {
		return sm, err
	}

	// 檢查模塊是否已經存在
	ok, err := moduleExists(s.path, moduleName)
	if err != nil {
		return sm, err
	}
	if ok {
		return sm, fmt.Errorf("模塊 %v 已經存在", moduleName)
	}

	// 應用選項
	var creationOpts moduleCreationOptions
	for _, apply := range options {
		apply(&creationOpts)
	}

	// 使用關聯類型解析參數
	params, err := field.ParseFields(creationOpts.params, checkForbiddenTypeIndex)
	if err != nil {
		return sm, err
	}

	// 檢查依賴關係
	if err := checkDependencies(creationOpts.dependencies, s.path); err != nil {
		return sm, err
	}

	opts := &modulecreate.CreateOptions{
		ModuleName:   moduleName,
		ModulePath:   s.modpath.RawPath,
		Params:       params,
		AppName:      s.modpath.Package,
		AppPath:      s.path,
		IsIBC:        creationOpts.ibc,
		IBCOrdering:  creationOpts.ibcChannelOrdering,
		Dependencies: creationOpts.dependencies,
	}

	// 來自 Cosmos SDK 版本的生成器
	g, err := modulecreate.NewStargate(opts)
	if err != nil {
		return sm, err
	}
	gens := []*genny.Generator{g}

	// 腳手架 IBC 模塊
	if opts.IsIBC {
		g, err = modulecreate.NewIBC(tracer, opts)
		if err != nil {
			return sm, err
		}
		gens = append(gens, g)
	}
	sm, err = xgenny.RunWithValidation(tracer, gens...)
	if err != nil {
		return sm, err
	}

	// 修改 app.go 註冊模塊
	newSourceModification, runErr := xgenny.RunWithValidation(tracer, modulecreate.NewStargateAppModify(tracer, opts))
	sm.Merge(newSourceModification)
	var validationErr validation.Error
	if runErr != nil && !errors.As(runErr, &validationErr) {
		return sm, runErr
	}

	return sm, finish(cacheStorage, opts.AppPath, s.modpath.RawPath)
}

// ImportModule 將具有名稱的指定模塊導入腳手架應用程序。
func (s Scaffolder) ImportModule(cacheStorage cache.Storage, tracer *placeholder.Tracer, name string) (sm xgenny.SourceModification, err error) {
	// Only wasm is currently supported
	if name != "wasm" {
		return sm, errors.New("模塊無法導入。支持模塊：wasm")
	}

	ok, err := isWasmImported(s.path)
	if err != nil {
		return sm, err
	}
	if ok {
		return sm, errors.New("wasm已經導入")
	}

	// run generator
	g, err := moduleimport.NewStargate(tracer, &moduleimport.ImportOptions{
		AppPath:          s.path,
		Feature:          name,
		AppName:          s.modpath.Package,
		BinaryNamePrefix: s.modpath.Root,
	})
	if err != nil {
		return sm, err
	}

	sm, err = xgenny.RunWithValidation(tracer, g)
	if err != nil {
		var validationErr validation.Error
		if errors.As(err, &validationErr) {
			// TODO：當有新方法要導入時，實現更通用的方法wasm
			return sm, errors.New("wasm無法導入.使用熊網鏈初始化的應用程序 <=0.16.2必須將熊網鏈版本降級為 0.16.2 在導入wasm")
		}
		return sm, err
	}

	// 導入特定版本的 ComsWasm
	// 注意（dshulyak）它必須在驗證後安裝
	if err := s.installWasm(); err != nil {
		return sm, err
	}

	return sm, finish(cacheStorage, s.path, s.modpath.RawPath)
}

// moduleExists 檢查模塊是否存在於應用程序中
func moduleExists(appPath string, moduleName string) (bool, error) {
	absPath, err := filepath.Abs(filepath.Join(appPath, moduleDir, moduleName))
	if err != nil {
		return false, err
	}

	_, err = os.Stat(absPath)
	if os.IsNotExist(err) {
		// 模塊不存在
		return false, nil
	}

	return err == nil, err
}

// checkModuleName 檢查名稱是否可以用作模塊名稱
func checkModuleName(appPath, moduleName string) error {
	// 去關鍵字
	if token.Lookup(moduleName).IsKeyword() {
		return fmt.Errorf("%s 是一個 Go 關鍵字", moduleName)
	}

	// 檢查名稱是否為保留名稱
	if _, ok := reservedNames[moduleName]; ok {
		return fmt.Errorf("%s 是保留名稱，不能用作模塊名稱", moduleName)
	}

	checkPrefix := func(name, prefix string) error {
		if strings.HasPrefix(name, prefix) {
			return fmt.Errorf("模塊名稱不能以 %s 因為潛在的存儲鍵衝突", prefix)
		}
		return nil
	}

	// 檢查名稱是否暗示潛在的存儲鍵衝突
	for _, defaultStoreKey := range defaultStoreKeys {
		if err := checkPrefix(moduleName, defaultStoreKey); err != nil {
			return err
		}
	}

	// 使用用戶定義的模塊檢查存儲鍵
	// 我們認為所有用戶定義的模塊都使用模塊名稱作為存儲鍵
	entries, err := os.ReadDir(filepath.Join(appPath, moduleDir))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if err := checkPrefix(moduleName, entry.Name()); err != nil {
			return err
		}
	}

	return nil
}

func isWasmImported(appPath string) (bool, error) {
	abspath := filepath.Join(appPath, appPkg)
	fset := token.NewFileSet()
	all, err := parser.ParseDir(fset, abspath, func(os.FileInfo) bool { return true }, parser.ImportsOnly)
	if err != nil {
		return false, err
	}
	for _, pkg := range all {
		for _, f := range pkg.Files {
			for _, imp := range f.Imports {
				if strings.Contains(imp.Path.Value, wasmImport) {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func (s Scaffolder) installWasm() error {
	switch {
	case s.Version.GTE(cosmosver.StargateFortyVersion):
		return cmdrunner.
			New().
			Run(context.Background(),
				step.New(step.Exec(gocmd.Name(), "get", gocmd.PackageLiteral(wasmImport, wasmVersion))),
				step.New(step.Exec(gocmd.Name(), "get", gocmd.PackageLiteral(extrasImport, extrasVersion))),
			)
	default:
		return errors.New("不支持的版本")
	}
}

//checkDependencies 對依賴項執行檢查
func checkDependencies(dependencies []modulecreate.Dependency, appPath string) error {
	depMap := make(map[string]struct{})
	for _, dep := range dependencies {
		//檢查依賴項是否已註冊
		path := filepath.Join(appPath, module.PathAppModule)
		if err := appanalysis.CheckKeeper(path, dep.KeeperName); err != nil {
			return fmt.Errorf(
				"模塊不能有 %s 作為依賴: %s",
				dep.Name,
				err.Error(),
			)
		}

		// 檢查重複
		_, ok := depMap[dep.Name]
		if ok {
			return fmt.Errorf("%s 是一個重複的依賴", dep)
		}
		depMap[dep.Name] = struct{}{}
	}

	return nil
}
