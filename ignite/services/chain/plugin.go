package chain

import (
	"context"

	"github.com/ignite-hq/cli/ignite/chainconfig"
	chaincmdrunner "github.com/ignite-hq/cli/ignite/pkg/chaincmd/runner"
)

// TODO 省略 Stargate 的 -cli 日誌消息。

type Plugin interface {
	// Cosmos 版本的名稱。
	Name() string

	// Gentx 返回 gentx 命令的 step.Exec 配置。
	Gentx(context.Context, chaincmdrunner.Runner, Validator) (path string, err error)

	// Configure 配置默認值.
	Configure(string, chainconfig.Config) error

	// Start 返回 step.Exec 配置以啟動服務器。
	Start(context.Context, chaincmdrunner.Runner, chainconfig.Config) error

	// Home 返回區塊鏈節點的主目錄。
	Home() string
}

func (c *Chain) pickPlugin() Plugin {
	return newStargatePlugin(c.app)
}
