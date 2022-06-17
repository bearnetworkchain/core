package ignitecmd

import (
	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/pkg/cliui"
	"github.com/bearnetworkchain/core/ignite/services/network"
)

const flagOut = "out"

// NewNetworkChainShow 創建一個新的鏈秀
// 命令顯示 SPN 上的鏈詳細信息。
func NewNetworkChainShow() *cobra.Command {
	c := &cobra.Command{
		Use:   "show",
		Short: "顯示熊網鏈的詳細信息",
	}
	c.AddCommand(
		newNetworkChainShowInfo(),
		newNetworkChainShowGenesis(),
		newNetworkChainShowAccounts(),
		newNetworkChainShowValidators(),
		newNetworkChainShowPeers(),
	)
	return c
}

func networkChainLaunch(cmd *cobra.Command, args []string, session cliui.Session) (NetworkBuilder, uint64, error) {
	nb, err := newNetworkBuilder(cmd, CollectEvents(session.EventBus()))
	if err != nil {
		return nb, 0, err
	}
	// parse launch ID.
	launchID, err := network.ParseID(args[0])
	if err != nil {
		return nb, launchID, err
	}
	return nb, launchID, err
}
