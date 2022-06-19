// Package tsproto 提供對 protoc-gen-ts_proto protoc 插件的訪問。
package tsproto

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ignite-hq/cli/ignite/pkg/nodetime"
)

const pluginName = "protoc-gen-ts_proto"

// BinaryPath 返回 ts-proto 插件的二進製文件的路徑，因此可以將其傳遞給
// 通過 --plugin 選項的協議。
//
// protoc 對其插件的二進制名稱非常挑剔。對於 ts-proto，二進制名稱
// 將是 protoc-gen-ts_proto。
// 看看為什麼：https://github.com/stephenh/ts-proto/blob/7f76c05/README.markdown#quickstart。
func BinaryPath() (path string, cleanup func(), err error) {
	var command []string

	command, cleanup, err = nodetime.Command(nodetime.CommandTSProto)
	if err != nil {
		return
	}

	tmpdir := os.TempDir()
	path = filepath.Join(tmpdir, pluginName)

	// 通過為插件的二進製文件提供 protoc-gen-ts_proto 名稱來安慰 protoc。
	script := fmt.Sprintf(`#!/bin/bash
%s "$@"
`, strings.Join(command, " "))

	err = os.WriteFile(path, []byte(script), 0755)

	return
}
