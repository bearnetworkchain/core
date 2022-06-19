package ignitecmd

import "github.com/spf13/cobra"

// NewChain 返回一個命令，該命令將與編譯、服務相關的子命令分組
// 區塊鍊等等。
func NewChain() *cobra.Command {
	c := &cobra.Command{
		Use:     "chain [command]",
		Short:   "構建、初始化和啟動熊網區塊鏈節點或在熊網區塊鏈上執行其他操作",
		Long:    `構建、初始化和啟動熊網區塊鏈節點或在熊網區塊鏈上執行其他操作.`,
		Aliases: []string{"c"},
		Args:    cobra.ExactArgs(1),
	}

	c.AddCommand(
		NewChainServe(),
		NewChainBuild(),
		NewChainInit(),
		NewChainFaucet(),
		NewChainSimulate(),
	)

	return c
}
