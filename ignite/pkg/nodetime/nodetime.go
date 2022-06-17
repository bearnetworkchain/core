// Package nodetime 提供一個單一的、獨立的 NodeJS 運行時可執行文件，其中包含
// 幾個 NodeJS CLI 程序捆綁在其中，可以通過子命令訪問這些程序。
// CLI 捆綁的程序是 Ignite CLI 需要的，可以根據需要添加更多。
package nodetime

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"sync"

	"github.com/bearnetworkchain/core/ignite/pkg/localfs"
	"github.com/bearnetworkchain/core/ignite/pkg/nodetime/data"
)

// 包含的 CLI 列表。
const (
	// CommandTSProto 是 https://github.com/stephenh/ts-proto。
	CommandTSProto CommandName = "ts-proto"

	// CommandSTA 是 https://github.com/acacode/swagger-typescript-api。
	CommandSTA CommandName = "sta"

	// CommandSwaggerCombine 是 https://www.npmjs.com/package/swagger-combine。
	CommandSwaggerCombine CommandName = "swagger-combine"

	// CommandIBCSetup 是 https://github.com/confio/ts-relayer/blob/main/spec/ibc-setup.md。
	CommandIBCSetup = "ibc-setup"

	// CommandIBCRelayer 是 https://github.com/confio/ts-relayer/blob/main/spec/ibc-relayer.md。
	CommandIBCRelayer = "ibc-relayer"

	// CommandXRelayer 是使用 confio 中繼器製作的 Ignite CLI 的中繼器包裝器。
	CommandXRelayer = "xrelayer"
)

// CommandName 表示 nodetime 下的高級命令。
type CommandName string

var (
	onceBinary sync.Once
	binary     []byte
)

// Binary 返回可執行文件的二進製字節。
func Binary() []byte {
	onceBinary.Do(func() {
		// 解壓二進製文件。
		gzr, err := gzip.NewReader(bytes.NewReader(data.Binary()))
		if err != nil {
			panic(err)
		}
		defer gzr.Close()

		tr := tar.NewReader(gzr)

		if _, err := tr.Next(); err != nil {
			panic(err)
		}

		if binary, err = io.ReadAll(tr); err != nil {
			panic(err)
		}
	})

	return binary
}

// 命令設置 nodetime 二進製文件並返回執行 c 所需的命令。
func Command(c CommandName) (command []string, cleanup func(), err error) {
	cs := string(c)
	path, cleanup, err := localfs.SaveBytesTemp(Binary(), cs, 0755)
	if err != nil {
		return nil, nil, err
	}
	command = []string{
		path,
		cs,
	}
	return command, cleanup, nil
}
