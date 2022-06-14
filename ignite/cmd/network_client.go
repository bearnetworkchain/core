package ignitecmd

import (
	"github.com/spf13/cobra"
)

// NewNetworkClient 創建一個新的客戶端命令，該命令包含其他一些
// 與將客戶端連接到網絡相關的子命令.
func NewNetworkClient() *cobra.Command {
	c := &cobra.Command{
		Use:   "client",
		Short: "使用 SPN 連接您的網絡",
	}

	c.AddCommand(
		NewNetworkClientCreate(),
	)

	return c
}
