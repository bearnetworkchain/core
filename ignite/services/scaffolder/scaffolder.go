// Package scaffolder 初始化 Ignite CLI 應用程序並修改現有應用程序
// 稍後添加更多功能。
package scaffolder

import (
	"context"
	"os"
	"path/filepath"

	"github.com/bearnetworkchain/core/ignite/chainconfig"
	sperrors "github.com/bearnetworkchain/core/ignite/errors"
	"github.com/bearnetworkchain/core/ignite/pkg/cache"
	"github.com/bearnetworkchain/core/ignite/pkg/cmdrunner"
	"github.com/bearnetworkchain/core/ignite/pkg/cmdrunner/step"
	"github.com/bearnetworkchain/core/ignite/pkg/cosmosanalysis"
	"github.com/bearnetworkchain/core/ignite/pkg/cosmosgen"
	"github.com/bearnetworkchain/core/ignite/pkg/cosmosver"
	"github.com/bearnetworkchain/core/ignite/pkg/gocmd"
	"github.com/bearnetworkchain/core/ignite/pkg/gomodule"
	"github.com/bearnetworkchain/core/ignite/pkg/gomodulepath"
)

// Scaffolder是 Ignite CLI 應用程序腳手架。
type Scaffolder struct {
	// 鏈的版本
	Version cosmosver.Version

	// 應用程序的路徑。
	path string

	// modpath表示應用的 go 模塊路徑。
	modpath gomodulepath.Path
}

// App為現有應用程序創建一個新的腳手架。
func App(path string) (Scaffolder, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return Scaffolder{}, err
	}

	modpath, path, err := gomodulepath.Find(path)
	if err != nil {
		return Scaffolder{}, err
	}
	modfile, err := gomodule.ParseAt(path)
	if err != nil {
		return Scaffolder{}, err
	}
	if err := cosmosanalysis.ValidateGoMod(modfile); err != nil {
		return Scaffolder{}, err
	}

	version, err := cosmosver.Detect(path)
	if err != nil {
		return Scaffolder{}, err
	}

	if !version.IsFamily(cosmosver.Stargate) {
		return Scaffolder{}, sperrors.ErrOnlyStargateSupported
	}

	s := Scaffolder{
		Version: version,
		path:    path,
		modpath: modpath,
	}

	return s, nil
}

func finish(cacheStorage cache.Storage, path, gomodPath string) error {
	if err := protoc(cacheStorage, path, gomodPath); err != nil {
		return err
	}
	if err := tidy(path); err != nil {
		return err
	}
	return fmtProject(path)
}

func protoc(cacheStorage cache.Storage, projectPath, gomodPath string) error {
	if err := cosmosgen.InstallDependencies(context.Background(), projectPath); err != nil {
		return err
	}

	confpath, err := chainconfig.LocateDefault(projectPath)
	if err != nil {
		return err
	}
	conf, err := chainconfig.ParseFile(confpath)
	if err != nil {
		return err
	}

	options := []cosmosgen.Option{
		cosmosgen.WithGoGeneration(gomodPath),
		cosmosgen.IncludeDirs(conf.Build.Proto.ThirdPartyPaths),
	}

	// generate 如果啟用了 Vuex 代碼，也是如此。
	if conf.Client.Vuex.Path != "" {
		storeRootPath := filepath.Join(projectPath, conf.Client.Vuex.Path, "generated")

		options = append(options,
			cosmosgen.WithVuexGeneration(
				false,
				cosmosgen.VuexStoreModulePath(storeRootPath),
				storeRootPath,
			),
		)
	}
	if conf.Client.OpenAPI.Path != "" {
		options = append(options, cosmosgen.WithOpenAPIGeneration(conf.Client.OpenAPI.Path))
	}

	return cosmosgen.Generate(context.Background(), cacheStorage, projectPath, conf.Build.Proto.Path, options...)
}

func tidy(path string) error {
	return cmdrunner.
		New(
			cmdrunner.DefaultStderr(os.Stderr),
			cmdrunner.DefaultWorkdir(path),
		).
		Run(context.Background(),
			step.New(
				step.Exec(gocmd.Name(), "mod", "tidy"),
			),
		)
}

func fmtProject(path string) error {
	return cmdrunner.
		New(
			cmdrunner.DefaultStderr(os.Stderr),
			cmdrunner.DefaultWorkdir(path),
		).
		Run(context.Background(),
			step.New(
				step.Exec(
					gocmd.Name(),
					"fmt",
					"./...",
				),
			),
		)
}
