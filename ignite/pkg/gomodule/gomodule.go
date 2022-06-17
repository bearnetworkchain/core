package gomodule

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"

	"github.com/bearnetworkchain/core/ignite/pkg/cache"
	"github.com/bearnetworkchain/core/ignite/pkg/cmdrunner"
	"github.com/bearnetworkchain/core/ignite/pkg/cmdrunner/step"
)

const pathCacheNamespace = "gomodule.path"

// ErrGoModNotFound 找不到應用的 go.mod 文件時返回。
var ErrGoModNotFound = errors.New("go.mod not found")

// ParseAt 查找和解析go.mod在應用程序的路徑。
func ParseAt(path string) (*modfile.File, error) {
	gomod, err := os.ReadFile(filepath.Join(path, "go.mod"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrGoModNotFound
		}
		return nil, err
	}
	return modfile.Parse("", gomod, nil)
}

// FilterVersions 通過路徑過濾 require 部分下的依賴項。
func FilterVersions(dependencies []module.Version, paths ...string) []module.Version {
	var filtered []module.Version

	for _, dep := range dependencies {
		for _, path := range paths {
			if dep.Path == path {
				filtered = append(filtered, dep)
				break
			}
		}
	}

	return filtered
}

func ResolveDependencies(f *modfile.File) ([]module.Version, error) {
	var versions []module.Version

	isReplacementAdded := func(rv module.Version) bool {
		for _, rep := range f.Replace {
			if rv.Path == rep.Old.Path {
				versions = append(versions, rep.New)

				return true
			}
		}

		return false
	}

	for _, req := range f.Require {
		if req.Indirect {
			continue
		}
		if !isReplacementAdded(req.Mod) {
			versions = append(versions, req.Mod)
		}
	}

	return versions, nil
}

// LocatePath 在本地文件系統上定位由 'go mod' 管理的 pkg 的絕對路徑。
func LocatePath(ctx context.Context, cacheStorage cache.Storage, src string, pkg module.Version) (path string, err error) {
	//可以是本地包。
	if pkg.Version == "" { // 表示這是本地包。
		if filepath.IsAbs(pkg.Path) {
			return pkg.Path, nil
		}
		return filepath.Join(src, pkg.Path), nil
	}

	pathCache := cache.New[string](cacheStorage, pathCacheNamespace)
	cacheKey := cache.Key(pkg.Path, pkg.Version)
	path, err = pathCache.Get(cacheKey)
	if err != nil && err != cache.ErrorNotFound {
		return "", err
	}
	if err != cache.ErrorNotFound {
		return path, nil
	}

	// 否則，它是託管的。
	out := &bytes.Buffer{}

	if err := cmdrunner.
		New().
		Run(ctx, step.New(
			step.Exec("go", "mod", "download", "-json"),
			step.Workdir(src),
			step.Stdout(out),
		)); err != nil {
		return "", err
	}

	d := json.NewDecoder(out)

	for {
		var module struct {
			Path, Version, Dir string
		}
		if err := d.Decode(&module); err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		if module.Path == pkg.Path && module.Version == pkg.Version {
			if err := pathCache.Put(cacheKey, module.Dir); err != nil {
				return "", err
			}
			return module.Dir, nil
		}
	}

	return "", fmt.Errorf("模塊 %q 未找到", pkg.Path)
}
