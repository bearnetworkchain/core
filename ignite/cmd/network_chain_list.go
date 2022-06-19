package ignitecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/pkg/cliui/entrywriter"
	"github.com/ignite-hq/cli/ignite/services/network"
	"github.com/ignite-hq/cli/ignite/services/network/networktypes"
)

var LaunchSummaryHeader = []string{"launch ID", "chain ID", "source", "campaign ID", "network", "reward"}

// NewNetworkChainList 返回一個新命令來列出 Ignite 上所有已發布的鏈
func NewNetworkChainList() *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "列出已發布的鏈",
		Args:  cobra.NoArgs,
		RunE:  networkChainListHandler,
	}
	return c
}

func networkChainListHandler(cmd *cobra.Command, _ []string) error {
	session := cliui.New()
	defer session.Cleanup()

	nb, err := newNetworkBuilder(cmd, CollectEvents(session.EventBus()))
	if err != nil {
		return err
	}
	n, err := nb.Network(network.CollectEvents(session.EventBus()))
	if err != nil {
		return err
	}
	chainLaunches, err := n.ChainLaunchesWithReward(cmd.Context())
	if err != nil {
		return err
	}

	session.StopSpinner()

	return renderLaunchSummaries(chainLaunches, session)
}

// renderLaunchSummaries 寫入提供的輸出，匯總啟動列表
func renderLaunchSummaries(chainLaunches []networktypes.ChainLaunch, session cliui.Session) error {
	var launchEntries [][]string

	for _, c := range chainLaunches {
		campaign := "沒有活動"
		if c.CampaignID > 0 {
			campaign = fmt.Sprintf("%d", c.CampaignID)
		}

		reward := entrywriter.None
		if len(c.Reward) > 0 {
			reward = c.Reward
		}

		launchEntries = append(launchEntries, []string{
			fmt.Sprintf("%d", c.ID),
			c.ChainID,
			c.SourceURL,
			campaign,
			c.Network.String(),
			reward,
		})
	}

	return session.PrintTable(LaunchSummaryHeader, launchEntries...)
}
