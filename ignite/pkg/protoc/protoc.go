// Package protoc 提供對 protoc 命令的高級訪問。
package protoc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/exec"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"
	"github.com/ignite-hq/cli/ignite/pkg/localfs"
	"github.com/ignite-hq/cli/ignite/pkg/protoanalysis"
	"github.com/ignite-hq/cli/ignite/pkg/protoc/data"
)

// Option配置 生成配置。
type Option func(*configs)

// configs持有生成配置。
type configs struct {
	pluginPath             string
	isGeneratedDepsEnabled bool
	pluginOptions          []string
	env                    []string
}

// Plugin配置一個用於代碼生成的插件。
func Plugin(path string, options ...string) Option {
	return func(c *configs) {
		c.pluginPath = path
		c.pluginOptions = options
	}
}

// Generate Dependencies 為您的 proto 文件所依賴的 proto 文件啟用代碼生成。
// 如果您的 protoc 插件沒有為您提供啟用相同功能的選項，請使用此選項。
func GenerateDependencies() Option {
	return func(c *configs) {
		c.isGeneratedDepsEnabled = true
	}
}

// Env 在代碼生成期間分配環境值。
func Env(v ...string) Option {
	return func(c *configs) {
		c.env = v
	}
}

type Cmd struct {
	Command  []string
	Included []string
}

// Command 設置 protoc 二進製文件並返回執行 c 所需的命令。
func Command() (command Cmd, cleanup func(), err error) {
	path, cleanupProto, err := localfs.SaveBytesTemp(data.Binary(), "protoc", 0755)
	if err != nil {
		return Cmd{}, nil, err
	}

	include, cleanupInclude, err := localfs.SaveTemp(data.Include())
	if err != nil {
		cleanupProto()
		return Cmd{}, nil, err
	}

	cleanup = func() {
		cleanupProto()
		cleanupInclude()
	}

	command = Cmd{
		Command:  []string{path, "-I", include},
		Included: []string{include},
	}

	return command, cleanup, nil
}

// Generate 使用 protocOuts 提供的插件從 protoPath 及其 includePaths 生成代碼到 outDir。
func Generate(ctx context.Context, outDir, protoPath string, includePaths, protocOuts []string, options ...Option) error {
	c := configs{}

	for _, o := range options {
		o(&c)
	}

	cmd, cleanup, err := Command()
	if err != nil {
		return err
	}
	defer cleanup()

	command := cmd.Command

	// 如果設置添加插件。
	if c.pluginPath != "" {
		command = append(command, "--plugin", c.pluginPath)
	}
	var existentIncludePaths []string

	//如果文件系統上實際上不存在第三方 proto 源，請跳過。
	for _, path := range includePaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		existentIncludePaths = append(existentIncludePaths, path)
	}

	// 將第三方原型位置附加到命令中。
	for _, importPath := range existentIncludePaths {
		command = append(command, "-I", importPath)
	}

	// 找出要為其生成代碼並執行代碼生成的 proto 文件列表。
	files, err := discoverFiles(ctx, c, protoPath, append(cmd.Included, existentIncludePaths...), protoanalysis.NewCache())
	if err != nil {
		return err
	}

	// 為每個 protocOuts 運行命令。
	for _, out := range protocOuts {
		command := append(command, out)
		command = append(command, files...)
		command = append(command, c.pluginOptions...)

		execOpts := []exec.Option{
			exec.StepOption(step.Workdir(outDir)),
			exec.IncludeStdLogsToError(),
		}
		if c.env != nil {
			execOpts = append(execOpts, exec.StepOption(step.Env(c.env...)))
		}

		if err := exec.Exec(ctx, command, execOpts...); err != nil {
			return err
		}
	}

	return nil
}

// discoverFiles 發現要為其生成代碼的 .proto 文件。應用程序的 .proto 文件
//（protoPath 下的所有內容）將始終是已發現文件的一部分。
//
// 當應用的 .proto 文件依賴於 includePaths (dependencies) 下的另一個 proto 包時，那些
// 也可能需要發現。一些 protoc 插件已經在內部進行了這個發現，但是
// 對於沒有的，如果啟用了 GenerateDependencies() 則需要在這里處理。
func discoverFiles(ctx context.Context, c configs, protoPath string, includePaths []string, cache protoanalysis.Cache) (
	discovered []string, err error) {
	packages, err := protoanalysis.Parse(ctx, cache, protoPath)
	if err != nil {
		return nil, err
	}

	discovered = packages.Files().Paths()

	if !c.isGeneratedDepsEnabled {
		return discovered, nil
	}

	for _, file := range packages.Files() {
		d, err := searchFile(file, protoPath, includePaths)
		if err != nil {
			return nil, err
		}
		discovered = append(discovered, d...)
	}

	return discovered, nil
}

func searchFile(file protoanalysis.File, protoPath string, includePaths []string) (discovered []string, err error) {
	dir := filepath.Dir(file.Path)

	for _, dep := range file.Dependencies {
		// 嘗試相對於這個 .proto 文件定位導入的 .proto 文件。
		guessedPath := filepath.Join(dir, dep)
		_, err := os.Stat(guessedPath)
		if err == nil {
			discovered = append(discovered, guessedPath)
			continue
		}
		if !os.IsNotExist(err) {
			return nil, err
		}

		// 否則，在 includePaths 中按絕對路徑搜索。
		var found bool
		for _, included := range includePaths {
			guessedPath := filepath.Join(included, dep)
			_, err := os.Stat(guessedPath)
			if err == nil {
// 找到依賴。
// 如果它在 protoPath 下，它已經被發現，所以跳過它。
				if !strings.HasPrefix(guessedPath, protoPath) {
					discovered = append(discovered, guessedPath)

					// 對這個執行完整的搜索以發現它的依賴關係。
					depFile, err := protoanalysis.ParseFile(guessedPath)
					if err != nil {
						return nil, err
					}
					d, err := searchFile(depFile, protoPath, includePaths)
					if err != nil {
						return nil, err
					}
					discovered = append(discovered, d...)
				}

				found = true
				break
			}
			if !os.IsNotExist(err) {
				return nil, err
			}
		}

		if !found {
			return nil, fmt.Errorf("找不到依賴項 %q 給予 %q", dep, file.Path)
		}
	}

	return discovered, nil
}
