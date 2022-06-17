package gocmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bearnetworkchain/core/ignite/pkg/cmdrunner/exec"
	"github.com/bearnetworkchain/core/ignite/pkg/cmdrunner/step"
	"github.com/bearnetworkchain/core/ignite/pkg/goenv"
)

const (
	// CommandInstall 表示執行“安裝”命令。
	CommandInstall = "install"

	// CommandBuild 代表 go "build" 命令。
	CommandBuild = "build"

	// CommandMod 代表 go "mod" 命令。
	CommandMod = "mod"

	// CommandModTidy 代表 go mod "tidy" 命令。
	CommandModTidy = "tidy"

	// CommandModVerify 代表 go mod "verify" 命令。
	CommandModVerify = "verify"
)

const (
	FlagMod              = "-mod"
	FlagModValueReadOnly = "readonly"
	FlagLdflags          = "-ldflags"
	FlagOut              = "-o"
)

const (
	EnvGOOS   = "GOOS"
	EnvGOARCH = "GOARCH"
)

// Name 返回要使用的 Go 二進製文件的名稱。
func Name() string {
	custom := os.Getenv("GONAME")
	if custom != "" {
		return custom
	}
	return "go"
}

// ModTidy 在帶有選項的路徑上運行 go mod tidy。
func ModTidy(ctx context.Context, path string, options ...exec.Option) error {
	return exec.Exec(ctx, []string{Name(), CommandMod, CommandModTidy}, append(options, exec.StepOption(step.Workdir(path)))...)
}

// ModVerify 在帶有選項的路徑上運行 go mod verify。
func ModVerify(ctx context.Context, path string, options ...exec.Option) error {
	return exec.Exec(ctx, []string{Name(), CommandMod, CommandModVerify}, append(options, exec.StepOption(step.Workdir(path)))...)
}

// BuildPath 在帶有選項的 cmd 文件夾上運行 go install。
func BuildPath(ctx context.Context, output, binary, path string, flags []string, options ...exec.Option) error {
	binaryOutput, err := binaryPath(output, binary)
	if err != nil {
		return err
	}
	command := []string{
		Name(),
		CommandBuild,
		FlagOut, binaryOutput,
	}
	command = append(command, flags...)
	command = append(command, ".")
	return exec.Exec(ctx, command, append(options, exec.StepOption(step.Workdir(path)))...)
}

// BuildAll 在帶有選項的路徑上運行 go build ./...。
func BuildAll(ctx context.Context, out, path string, flags []string, options ...exec.Option) error {
	command := []string{
		Name(),
		CommandBuild,
		FlagOut, out,
	}
	command = append(command, flags...)
	command = append(command, "./...")
	return exec.Exec(ctx, command, append(options, exec.StepOption(step.Workdir(path)))...)
}

// InstallAll 在帶有選項的路徑上運行 go install ./...。
func InstallAll(ctx context.Context, path string, flags []string, options ...exec.Option) error {
	command := []string{
		Name(),
		CommandInstall,
	}
	command = append(command, flags...)
	command = append(command, "./...")
	return exec.Exec(ctx, command, append(options, exec.StepOption(step.Workdir(path)))...)
}

// Ldflags 返回一個組合ldflags 從設置 flags.
func Ldflags(flags ...string) string {
	return strings.Join(flags, " ")
}

// BuildTarget 構建一個 GOOS:GOARCH 對。
func BuildTarget(goos, goarch string) string {
	return fmt.Sprintf("%s:%s", goos, goarch)
}

// ParseTarget 解析 GOOS:GOARCH 對。
func ParseTarget(t string) (goos, goarch string, err error) {
	parsed := strings.Split(t, ":")
	if len(parsed) != 2 {
		return "", "", errors.New("無效的 Go 目標，預期為 GOOS:GOARCH 格式")
	}

	return parsed[0], parsed[1], nil
}

// PackageLiteral 返回 go get 的包部分的字符串表示[package].
func PackageLiteral(path, version string) string {
	return fmt.Sprintf("%s@%s", path, version)
}

// binaryPath 確定二進製文件所在的路徑。
func binaryPath(output, binary string) (string, error) {
	if output != "" {
		outputAbs, err := filepath.Abs(output)
		if err != nil {
			return "", err
		}
		return filepath.Join(outputAbs, binary), nil
	}
	return filepath.Join(goenv.Bin(), binary), nil
}
