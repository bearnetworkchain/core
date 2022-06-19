package ignitecmd

import (
	"github.com/spf13/cobra"
)

// NewNetworkChain 創建一個新的鏈命令，其中包含一些其他的
// 與為鏈啟動網絡相關的子命令。
func NewNetworkChain() *cobra.Command {
	c := &cobra.Command{
		Use:   "chain",
		Short: "建立網絡",
	}

	c.AddCommand(
		NewNetworkChainList(),
		NewNetworkChainPublish(),
		NewNetworkChainInit(),
		NewNetworkChainInstall(),
		NewNetworkChainJoin(),
		NewNetworkChainPrepare(),
		NewNetworkChainShow(),
		NewNetworkChainLaunch(),
		NewNetworkChainRevertLaunch(),
	)

	return c
}
