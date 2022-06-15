package chain

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/docker/docker/pkg/archive"
	"github.com/pkg/errors"

	"github.com/ignite-hq/cli/ignite/pkg/cache"
	"github.com/ignite-hq/cli/ignite/pkg/checksum"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/exec"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"
	"github.com/ignite-hq/cli/ignite/pkg/dirchange"
	"github.com/ignite-hq/cli/ignite/pkg/goanalysis"
	"github.com/ignite-hq/cli/ignite/pkg/gocmd"
	"github.com/ignite-hq/cli/ignite/pkg/xstrings"
)

const (
	releaseDir                   = "release"
	releaseChecksumKey           = "release_checksum"
	modChecksumKey               = "go_mod_checksum"
	buildDirchangeCacheNamespace = "build.dirchange"
)

//Build 構建並安裝應用程序二進製文件。
func (c *Chain) Build(ctx context.Context, cacheStorage cache.Storage, output string) (binaryName string, err error) {
	if err := c.setup(); err != nil {
		return "", err
	}

	if err := c.build(ctx, cacheStorage, output); err != nil {
		return "", err
	}

	return c.Binary()
}

func (c *Chain) build(ctx context.Context, cacheStorage cache.Storage, output string) (err error) {
	defer func() {
		var exitErr *exec.ExitError

		if errors.As(err, &exitErr) || errors.Is(err, goanalysis.ErrMultipleMainPackagesFound) {
			err = &CannotBuildAppError{err}
		}
	}()

	if err := c.generateAll(ctx, cacheStorage); err != nil {
		return err
	}

	buildFlags, err := c.preBuild(ctx, cacheStorage)
	if err != nil {
		return err
	}

	binary, err := c.Binary()
	if err != nil {
		return err
	}

	path, err := c.discoverMain(c.app.Path)
	if err != nil {
		return err
	}

	return gocmd.BuildPath(ctx, output, binary, path, buildFlags)
}

// BuildRelease 為發布構建二進製文件。目標是一個列表
// GOOS:GOARCH 提供時。當沒有提供目標時，它默認為您的系統。
// 前綴用作包含每個目標的 tarball 的前綴。
func (c *Chain) BuildRelease(ctx context.Context, cacheStorage cache.Storage, output, prefix string, targets ...string) (releasePath string, err error) {
	if prefix == "" {
		prefix = c.app.Name
	}
	if len(targets) == 0 {
		targets = []string{gocmd.BuildTarget(runtime.GOOS, runtime.GOARCH)}
	}

	//準備構建。
	if err := c.setup(); err != nil {
		return "", err
	}

	buildFlags, err := c.preBuild(ctx, cacheStorage)
	if err != nil {
		return "", err
	}

	binary, err := c.Binary()
	if err != nil {
		return "", err
	}

	mainPath, err := c.discoverMain(c.app.Path)
	if err != nil {
		return "", err
	}

	releasePath = output
	if releasePath == "" {
		releasePath = filepath.Join(c.app.Path, releaseDir)
		// 重置發布目錄。
		if err := os.RemoveAll(releasePath); err != nil {
			return "", err
		}
	}

	if err := os.MkdirAll(releasePath, 0755); err != nil {
		return "", err
	}

	for _, t := range targets {
		//為目標構建二進製文件，將其壓縮並保存在發布目錄下。
		goos, goarch, err := gocmd.ParseTarget(t)
		if err != nil {
			return "", err
		}

		out, err := os.MkdirTemp("", "")
		if err != nil {
			return "", err
		}
		defer os.RemoveAll(out)

		buildOptions := []exec.Option{
			exec.StepOption(step.Env(
				cmdrunner.Env(gocmd.EnvGOOS, goos),
				cmdrunner.Env(gocmd.EnvGOARCH, goarch),
			)),
		}

		if err := gocmd.BuildPath(ctx, out, binary, mainPath, buildFlags, buildOptions...); err != nil {
			return "", err
		}

		tarr, err := archive.Tar(out, archive.Gzip)
		if err != nil {
			return "", err
		}

		tarName := fmt.Sprintf("%s_%s_%s.tar.gz", prefix, goos, goarch)
		tarPath := filepath.Join(releasePath, tarName)

		tarf, err := os.Create(tarPath)
		if err != nil {
			return "", err
		}
		defer tarf.Close()

		if _, err := io.Copy(tarf, tarr); err != nil {
			return "", err
		}
		tarf.Close()
	}

	checksumPath := filepath.Join(releasePath, releaseChecksumKey)

	// 創建一個 checksum.txt 並返回釋放目錄的路徑。
	return releasePath, checksum.Sum(releasePath, checksumPath)
}

func (c *Chain) preBuild(ctx context.Context, cacheStorage cache.Storage) (buildFlags []string, err error) {
	config, err := c.Config()
	if err != nil {
		return nil, err
	}

	chainID, err := c.ID()
	if err != nil {
		return nil, err
	}

	ldFlags := config.Build.LDFlags
	ldFlags = append(ldFlags,
		fmt.Sprintf("-X github.com/cosmos/cosmos-sdk/version.Name=%s", xstrings.Title(c.app.Name)),
		fmt.Sprintf("-X github.com/cosmos/cosmos-sdk/version.AppName=%sd", c.app.Name),
		fmt.Sprintf("-X github.com/cosmos/cosmos-sdk/version.Version=%s", c.sourceVersion.tag),
		fmt.Sprintf("-X github.com/cosmos/cosmos-sdk/version.Commit=%s", c.sourceVersion.hash),
		fmt.Sprintf("-X %s/cmd/%s/cmd.ChainID=%s", c.app.ImportPath, c.app.D(), chainID),
	)
	buildFlags = []string{
		gocmd.FlagMod, gocmd.FlagModValueReadOnly,
		gocmd.FlagLdflags, gocmd.Ldflags(ldFlags...),
	}

	fmt.Fprintln(c.stdLog().out, "📦 安裝熊網鏈依賴項...")

// 我們在檢查校驗和更改之前做 mod tidy，因為 go.mod 經常被修改
// 無論如何，mod verify 命令是昂貴的
	if err := gocmd.ModTidy(ctx, c.app.Path); err != nil {
		return nil, err
	}

	dirCache := cache.New[[]byte](cacheStorage, buildDirchangeCacheNamespace)
	modChanged, err := dirchange.HasDirChecksumChanged(dirCache, modChecksumKey, c.app.Path, "go.mod")
	if err != nil {
		return nil, err
	}

	if modChanged {
		if err := gocmd.ModVerify(ctx, c.app.Path); err != nil {
			return nil, err
		}

		if err := dirchange.SaveDirChecksum(dirCache, modChecksumKey, c.app.Path, "go.mod"); err != nil {
			return nil, err
		}
	}

	fmt.Fprintln(c.stdLog().out, "🛠️  構建熊網鏈...")

	return buildFlags, nil
}

func (c *Chain) discoverMain(path string) (pkgPath string, err error) {
	conf, err := c.Config()
	if err != nil {
		return "", err
	}

	if conf.Build.Main != "" {
		return filepath.Join(c.app.Path, conf.Build.Main), nil
	}

	path, err = goanalysis.DiscoverOneMain(path)
	if err == goanalysis.ErrMultipleMainPackagesFound {
		return "", errors.Wrap(err, "請在config.yml檔案中的build.main部份指定鏈主包的路徑")
	}
	return path, err
}
