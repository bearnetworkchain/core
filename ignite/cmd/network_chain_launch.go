package ignitecmd

import (
	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/services/network"
)

const (
	flagRemainingTime = "remaining-time"
)

// NewNetworkChainLaunch 創建一個新的鏈啟動命令來啟動
// 作為網絡協調者。
func NewNetworkChainLaunch() *cobra.Command {
	c := &cobra.Command{
		Use:   "launch [launch-id]",
		Short: "作為協調者啟動網絡",
		Args:  cobra.ExactArgs(1),
		RunE:  networkChainLaunchHandler,
	}

	c.Flags().Duration(flagRemainingTime, 0, "熊網鏈有效啟動前的持續時間（以秒為單位）")
	c.Flags().AddFlagSet(flagNetworkFrom())
	c.Flags().AddFlagSet(flagSetKeyringBackend())

	return c
}

func networkChainLaunchHandler(cmd *cobra.Command, args []string) error {
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

	remainingTime, _ := cmd.Flags().GetDuration(flagRemainingTime)

	n, err := nb.Network()
	if err != nil {
		return err
	}

	return n.TriggerLaunch(cmd.Context(), launchID, remainingTime)
}
