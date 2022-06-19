package ignitecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/pkg/cliui/entrywriter"
	"github.com/ignite-hq/cli/ignite/services/network/networktypes"
)

var CampaignSummaryHeader = []string{
	"id",
	"name",
	"coordinator id",
	"mainnet id",
}

// NewNetworkCampaignList 返回一個新命令以列出 Ignite 上所有已發布的活動。
func NewNetworkCampaignList() *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "列出已發布的活動",
		Args:  cobra.NoArgs,
		RunE:  networkCampaignListHandler,
	}
	return c
}

func networkCampaignListHandler(cmd *cobra.Command, _ []string) error {
	session := cliui.New()
	defer session.Cleanup()

	nb, err := newNetworkBuilder(cmd, CollectEvents(session.EventBus()))
	if err != nil {
		return err
	}

	n, err := nb.Network()
	if err != nil {
		return err
	}
	campaigns, err := n.Campaigns(cmd.Context())
	if err != nil {
		return err
	}

	return renderCampaignSummaries(campaigns, session)
}

//renderCampaignSummaries 寫入提供的總結活動列表
func renderCampaignSummaries(campaigns []networktypes.Campaign, session cliui.Session) error {
	var campaignEntries [][]string

	for _, c := range campaigns {
		mainnetID := entrywriter.None
		if c.MainnetInitialized {
			mainnetID = fmt.Sprintf("%d", c.MainnetID)
		}

		campaignEntries = append(campaignEntries, []string{
			fmt.Sprintf("%d", c.ID),
			c.Name,
			fmt.Sprintf("%d", c.CoordinatorID),
			mainnetID,
		})
	}

	session.StopSpinner()

	return session.PrintTable(CampaignSummaryHeader, campaignEntries...)
}
