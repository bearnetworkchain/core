package chain

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ignite-hq/cli/ignite/pkg/cache"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosanalysis/module"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosgen"
)

const (
	defaultVuexPath    = "vue/src/store"
	defaultDartPath    = "flutter/lib"
	defaultOpenAPIPath = "docs/static/openapi.yml"
)

type generateOptions struct {
	isGoEnabled      bool
	isVuexEnabled    bool
	isDartEnabled    bool
	isOpenAPIEnabled bool
}

// 生成目標是從 proto 文件生成代碼的目標。
type GenerateTarget func(*generateOptions)

// GenerateGo 可以生成鏈源代碼所需的基於 proto 的 Go 代碼。
func GenerateGo() GenerateTarget {
	return func(o *generateOptions) {
		o.isGoEnabled = true
	}
}

// GenerateVuex 可以生成基於 proto 的 Vuex 存儲。
func GenerateVuex() GenerateTarget {
	return func(o *generateOptions) {
		o.isVuexEnabled = true
	}
}

// GenerateDart 啟用生成 Dart 客戶端。
func GenerateDart() GenerateTarget {
	return func(o *generateOptions) {
		o.isDartEnabled = true
	}
}

// GenerateOpenAPI 可以為您的鏈生成 OpenAPI 規範。
func GenerateOpenAPI() GenerateTarget {
	return func(o *generateOptions) {
		o.isOpenAPIEnabled = true
	}
}

func (c *Chain) generateAll(ctx context.Context, cacheStorage cache.Storage) error {
	conf, err := c.Config()
	if err != nil {
		return err
	}

	var additionalTargets []GenerateTarget

	if conf.Client.Vuex.Path != "" {
		additionalTargets = append(additionalTargets, GenerateVuex())
	}

	if conf.Client.Dart.Path != "" {
		additionalTargets = append(additionalTargets, GenerateDart())
	}

	if conf.Client.OpenAPI.Path != "" {
		additionalTargets = append(additionalTargets, GenerateOpenAPI())
	}

	return c.Generate(ctx, cacheStorage, GenerateGo(), additionalTargets...)
}

// Generate 為給定的目標和附加目標從 proto 文件生成代碼。
func (c *Chain) Generate(
	ctx context.Context,
	cacheStorage cache.Storage,
	target GenerateTarget,
	additionalTargets ...GenerateTarget,
) error {
	var targetOptions generateOptions

	for _, apply := range append(additionalTargets, target) {
		apply(&targetOptions)
	}

	conf, err := c.Config()
	if err != nil {
		return err
	}

	if err := cosmosgen.InstallDependencies(ctx, c.app.Path); err != nil {
		return err
	}

	fmt.Fprintln(c.stdLog().out, "🛠️  構建原型...")

	options := []cosmosgen.Option{
		cosmosgen.IncludeDirs(conf.Build.Proto.ThirdPartyPaths),
	}

	if targetOptions.isGoEnabled {
		options = append(options, cosmosgen.WithGoGeneration(c.app.ImportPath))
	}

	enableThirdPartyModuleCodegen := !c.protoBuiltAtLeastOnce && c.options.isThirdPartyModuleCodegenEnabled

	// 如果啟用，也會生成 Vuex 代碼。
	if targetOptions.isVuexEnabled {
		vuexPath := conf.Client.Vuex.Path
		if vuexPath == "" {
			vuexPath = defaultVuexPath
		}

		storeRootPath := filepath.Join(c.app.Path, vuexPath, "generated")
		if err := os.MkdirAll(storeRootPath, 0766); err != nil {
			return err
		}

		options = append(options,
			cosmosgen.WithVuexGeneration(
				enableThirdPartyModuleCodegen,
				cosmosgen.VuexStoreModulePath(storeRootPath),
				storeRootPath,
			),
		)
	}

	if targetOptions.isDartEnabled {
		dartPath := conf.Client.Dart.Path

		if dartPath == "" {
			dartPath = defaultDartPath
		}

		rootPath := filepath.Join(c.app.Path, dartPath, "generated")
		if err := os.MkdirAll(rootPath, 0766); err != nil {
			return err
		}

		options = append(options,
			cosmosgen.WithDartGeneration(
				enableThirdPartyModuleCodegen,
				func(m module.Module) string {
					return filepath.Join(rootPath, m.Pkg.Name, "module")
				},
				rootPath,
			),
		)
	}

	if targetOptions.isOpenAPIEnabled {
		openAPIPath := conf.Client.OpenAPI.Path

		if openAPIPath == "" {
			openAPIPath = defaultOpenAPIPath
		}

		options = append(options, cosmosgen.WithOpenAPIGeneration(openAPIPath))
	}

	if err := cosmosgen.Generate(ctx, cacheStorage, c.app.Path, conf.Build.Proto.Path, options...); err != nil {
		return &CannotBuildAppError{err}
	}

	c.protoBuiltAtLeastOnce = true

	return nil
}
