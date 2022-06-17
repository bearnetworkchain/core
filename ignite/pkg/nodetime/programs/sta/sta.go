// Package sta 提供對 swagger-typescript-api CLI 的訪問。
package sta

import (
	"context"
	"path/filepath"

	"github.com/bearnetworkchain/core/ignite/pkg/cmdrunner/exec"
	"github.com/bearnetworkchain/core/ignite/pkg/nodetime"
)

// Generate 從位於 specPath 的 OpenAPI 規範生成客戶端代碼和 TS 類型到 outPath。
func Generate(ctx context.Context, outPath, specPath, moduleNameIndex string) error {
	command, cleanup, err := nodetime.Command(nodetime.CommandSTA)
	if err != nil {
		return err
	}
	defer cleanup()

	dir := filepath.Dir(outPath)
	file := filepath.Base(outPath)

	// command 構造 sta 命令。
	command = append(command, []string{
		"--module-name-index",
		moduleNameIndex,
		"-p",
		specPath,
		"-o",
		dir,
		"-n",
		file,
	}...)

	// 執行命令。
	return exec.Exec(ctx, command, exec.IncludeStdLogsToError())
}
