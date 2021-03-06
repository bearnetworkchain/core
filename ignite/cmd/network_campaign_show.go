package ignitecmd

import (
	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/pkg/yaml"
	"github.com/ignite-hq/cli/ignite/services/network"
)

// NewNetworkCampaignShow 返回一個新命令以在 Ignite 上顯示已發布的活動
func NewNetworkCampaignShow() *cobra.Command {
	c := &cobra.Command{
		Use:   "show [campaign-id]",
		Short: "顯示已發布的活動",
		Args:  cobra.ExactArgs(1),
		RunE:  networkCampaignShowHandler,
	}
	return c
}

func networkCampaignShowHandler(cmd *cobra.Command, args []string) error {
	session := cliui.New()
	defer session.Cleanup()

	// 解析活動 ID
	campaignID, err := network.ParseID(args[0])
	if err != nil {
		return err
	}

	nb, err := newNetworkBuilder(cmd, CollectEvents(session.EventBus()))
	if err != nil {
		return err
	}

	n, err := nb.Network()
	if err != nil {
		return err
	}

	campaign, err := n.Campaign(cmd.Context(), campaignID)
	if err != nil {
		return err
	}

	info, err := yaml.Marshal(cmd.Context(), campaign)
	if err != nil {
		return err
	}

	session.StopSpinner()

	return session.Println(info)
}
