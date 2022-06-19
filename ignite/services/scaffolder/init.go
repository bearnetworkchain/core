package scaffolder

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gobuffalo/genny"
	"github.com/tendermint/flutter/v2"
	"github.com/tendermint/vue"

	"github.com/ignite-hq/cli/ignite/pkg/cache"
	"github.com/ignite-hq/cli/ignite/pkg/gomodulepath"
	"github.com/ignite-hq/cli/ignite/pkg/localfs"
	"github.com/ignite-hq/cli/ignite/pkg/placeholder"
	"github.com/ignite-hq/cli/ignite/templates/app"
	modulecreate "github.com/ignite-hq/cli/ignite/templates/module/create"
)

var (
	commitMessage = "使用 Ignite CLI 初始化"
	devXAuthor    = &object.Signature{
		Name:  "熊網鏈的開發人員體驗團隊",
		Email: "bear.network.root@gmail.com",
		When:  time.Now(),
	}
)

// Init 使用名稱和給定選項初始化一個新應用程序。
func Init(cacheStorage cache.Storage, tracer *placeholder.Tracer, root, name, addressPrefix string, noDefaultModule bool) (path string, err error) {
	if root, err = filepath.Abs(root); err != nil {
		return "", err
	}

	pathInfo, err := gomodulepath.Parse(name)
	if err != nil {
		return "", err
	}

	path = filepath.Join(root, pathInfo.Root)

	// create the project
	if err := generate(tracer, pathInfo, addressPrefix, path, noDefaultModule); err != nil {
		return "", err
	}

	if err := finish(cacheStorage, path, pathInfo.RawPath); err != nil {
		return "", err
	}

	// 初始化 git 存儲庫並執行第一次提交
	if err := initGit(path); err != nil {
		return "", err
	}

	return path, nil
}

//不願意：界面
func generate(
	tracer *placeholder.Tracer,
	pathInfo gomodulepath.Path,
	addressPrefix,
	absRoot string,
	noDefaultModule bool,
) error {
	githubPath := gomodulepath.ExtractAppPath(pathInfo.RawPath)
	if !strings.Contains(githubPath, "/") {
		// 當應用模塊路徑只有一個元素時，必須添加用戶名
		githubPath = fmt.Sprintf("用戶名/%s", githubPath)
	}

	g, err := app.New(&app.Options{
		// 生成應用模板
		ModulePath:       pathInfo.RawPath,
		AppName:          pathInfo.Package,
		AppPath:          absRoot,
		GitHubPath:       githubPath,
		BinaryNamePrefix: pathInfo.Root,
		AddressPrefix:    addressPrefix,
	})
	if err != nil {
		return err
	}

	run := func(runner *genny.Runner, gen *genny.Generator) error {
		runner.With(gen)
		runner.Root = absRoot
		return runner.Run()
	}
	if err := run(genny.WetRunner(context.Background()), g); err != nil {
		return err
	}

	// 生成模塊模板
	if !noDefaultModule {
		opts := &modulecreate.CreateOptions{
			ModuleName: pathInfo.Package, //名稱
			ModulePath: pathInfo.RawPath,
			AppName:    pathInfo.Package,
			AppPath:    absRoot,
			IsIBC:      false,
		}
		g, err = modulecreate.NewStargate(opts)
		if err != nil {
			return err
		}
		if err := run(genny.WetRunner(context.Background()), g); err != nil {
			return err
		}
		g = modulecreate.NewStargateAppModify(tracer, opts)
		if err := run(genny.WetRunner(context.Background()), g); err != nil {
			return err
		}

	}

	// 生成 vue 應用。
	return Vue(filepath.Join(absRoot, "vue"))
}

//Vue 為鏈搭建了一個 Vue.js 應用程序。
func Vue(path string) error {
	return localfs.Save(vue.Boilerplate(), path)
}

// Flutter 為鏈構建了一個 Flutter 應用程序。
func Flutter(path string) error {
	return localfs.Save(flutter.Boilerplate(), path)
}

func initGit(path string) error {
	repo, err := git.PlainInit(path, false)
	if err != nil {
		return err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	if _, err := wt.Add("."); err != nil {
		return err
	}
	_, err = wt.Commit(commitMessage, &git.CommitOptions{
		All:    true,
		Author: devXAuthor,
	})
	return err
}
