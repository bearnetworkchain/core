package ignitecmd

import "github.com/spf13/cobra"

// NewChain returns a command that groups sub commands related to compiling, serving
// blockchains and so on.
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
