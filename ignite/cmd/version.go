package ignitecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/version"
)

// NewVersion 創建一個新版本命令來顯示 Ignite CLI 版本。
func NewVersion() *cobra.Command {
	c := &cobra.Command{
		Use:   "version",
		Short: "打印當前構建信息",
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Println(version.Long(cmd.Context()))
		},
	}
	return c
}
