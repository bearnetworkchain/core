package ignitecmd

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/services/network"
)

// NewNetworkRewardSet 創建一個新的鏈獎勵集命令以
// 將鏈獎勵作為協調者添加到網絡中。
func NewNetworkRewardSet() *cobra.Command {
	c := &cobra.Command{
		Use:   "set [launch-id] [last-reward-height] [coins]",
		Short: "設置網絡鏈獎勵",
		Args:  cobra.ExactArgs(3),
		RunE:  networkChainRewardSetHandler,
	}
	c.Flags().AddFlagSet(flagSetKeyringBackend())
	c.Flags().AddFlagSet(flagNetworkFrom())
	c.Flags().AddFlagSet(flagSetHome())
	return c
}

func networkChainRewardSetHandler(cmd *cobra.Command, args []string) error {
	session := cliui.New()
	defer session.Cleanup()

	nb, err := newNetworkBuilder(cmd, CollectEvents(session.EventBus()))
	if err != nil {
		return err
	}

	// 解析啟動 ID
	launchID, err := network.ParseID(args[0])
	if err != nil {
		return err
	}

	// 解析最後的獎勵高度
	lastRewardHeight, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return err
	}

	coins, err := sdk.ParseCoinsNormalized(args[2])
	if err != nil {
		return fmt.Errorf("無法解析硬幣: %w", err)
	}

	n, err := nb.Network()
	if err != nil {
		return err
	}

	return n.SetReward(launchID, lastRewardHeight, coins)
}
