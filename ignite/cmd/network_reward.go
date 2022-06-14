package ignitecmd

import (
	"github.com/spf13/cobra"
)

// NewNetworkReward 創建新的鏈獎勵命令
func NewNetworkReward() *cobra.Command {
	c := &cobra.Command{
		Use:   "reward",
		Short: "管理網絡獎勵",
	}
	c.AddCommand(
		NewNetworkRewardSet(),
	)
	return c
}
