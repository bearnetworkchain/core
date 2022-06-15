// Package goenv 定義 Go 已知的環境變量和一些圍繞它的實用程序。
package goenv

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
)

const (
	// GOBIN 是 GOBIN 的環境變量。
	GOBIN = "GOBIN"

	// GOPATH 是 GOPATH 的環境變量。
	GOPATH = "GOPATH"
)

const (
	binDir = "bin"
)

// Bin 返回安裝 Go 二進製文件的路徑。
func Bin() string {
	if binPath := os.Getenv(GOBIN); binPath != "" {
		return binPath
	}
	if goPath := os.Getenv(GOPATH); goPath != "" {
		return filepath.Join(goPath, binDir)
	}
	return filepath.Join(build.Default.GOPATH, binDir)
}

// Path 返回 $PATH 具有正確的 go bin 配置集。
func Path() string {
	return os.ExpandEnv(fmt.Sprintf("$PATH:%s", Bin()))
}

// ConfigurePath 使用具有 go bin 設置的正確 $PATH 配置 env。
func ConfigurePath() error {
	return os.Setenv("PATH", Path())
}
