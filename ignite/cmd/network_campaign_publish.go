package ignitecmd

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/pkg/cliui/icons"
)

const (
	flagMetadata = "metadata"
)

// NewNetworkCampaignPublish 返回一個用於在 Ignite 上發布新活動的新命令。
func NewNetworkCampaignPublish() *cobra.Command {
	c := &cobra.Command{
		Use:   "create [name] [total-supply]",
		Short: "創建一個活動",
		Args:  cobra.ExactArgs(2),
		RunE:  networkCampaignPublishHandler,
	}
	c.Flags().String(flagMetadata, "", "將元數據添加到鏈中")
	c.Flags().AddFlagSet(flagNetworkFrom())
	c.Flags().AddFlagSet(flagSetKeyringBackend())
	c.Flags().AddFlagSet(flagSetHome())
	return c
}

func networkCampaignPublishHandler(cmd *cobra.Command, args []string) error {
	session := cliui.New()
	defer session.Cleanup()

	nb, err := newNetworkBuilder(cmd, CollectEvents(session.EventBus()))
	if err != nil {
		return err
	}

	totalSupply, err := sdk.ParseCoinsNormalized(args[1])
	if err != nil {
		return err
	}

	n, err := nb.Network()
	if err != nil {
		return err
	}

	metadata, _ := cmd.Flags().GetString(flagMetadata)
	campaignID, err := n.CreateCampaign(args[0], metadata, totalSupply)
	if err != nil {
		return err
	}

	session.StopSpinner()

	return session.Printf("%s 活動ID: %d \n", icons.Bullet, campaignID)
}
