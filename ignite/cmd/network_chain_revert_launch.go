package ignitecmd

import (
	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/services/network"
	"github.com/ignite-hq/cli/ignite/services/network/networkchain"
)

// NewNetworkChainRevertLaunch creates a new chain revert launch command
// to revert a launched chain.
func NewNetworkChainRevertLaunch() *cobra.Command {
	c := &cobra.Command{
		Use:   "revert-launch [launch-id]",
		Short: "恢復以協調者身份啟動網絡",
		Args:  cobra.ExactArgs(1),
		RunE:  networkChainRevertLaunchHandler,
	}

	c.Flags().AddFlagSet(flagNetworkFrom())
	c.Flags().AddFlagSet(flagSetKeyringBackend())

	return c
}

func networkChainRevertLaunchHandler(cmd *cobra.Command, args []string) error {
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

	n, err := nb.Network()
	if err != nil {
		return err
	}

	chainLaunch, err := n.ChainLaunch(cmd.Context(), launchID)
	if err != nil {
		return err
	}

	c, err := nb.Chain(networkchain.SourceLaunch(chainLaunch))
	if err != nil {
		return err
	}

	return n.RevertLaunch(launchID, c)
}
