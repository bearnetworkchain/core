package ignitecmd

import (
	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/pkg/cliui/icons"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosclient"
	"github.com/ignite-hq/cli/ignite/services/network"
)

// NewNetworkClientCreate 將已啟動鏈的監控模塊與 SPN 連接起來
func NewNetworkClientCreate() *cobra.Command {
	c := &cobra.Command{
		Use:   "create [launch-id] [node-api-url]",
		Short: "將已啟動鏈的監控模塊與 SPN 連接",
		Args:  cobra.ExactArgs(2),
		RunE:  networkClientCreateHandler,
	}
	c.Flags().AddFlagSet(flagNetworkFrom())
	c.Flags().AddFlagSet(flagSetKeyringBackend())
	return c
}

func networkClientCreateHandler(cmd *cobra.Command, args []string) error {
	session := cliui.New()
	defer session.Cleanup()

	launchID, err := network.ParseID(args[0])
	if err != nil {
		return err
	}
	nodeAPI := args[1]

	nb, err := newNetworkBuilder(cmd)
	if err != nil {
		return err
	}

	nodeClient, err := cosmosclient.New(cmd.Context(), cosmosclient.WithNodeAddress(nodeAPI))
	if err != nil {
		return err
	}
	node, err := network.NewNodeClient(nodeClient)
	if err != nil {
		return err
	}

	rewardsInfo, unboundingTime, err := node.RewardsInfo(cmd.Context())
	if err != nil {
		return err
	}

	n, err := nb.Network()
	if err != nil {
		return err
	}

	clientID, err := n.CreateClient(launchID, unboundingTime, rewardsInfo)
	if err != nil {
		return err
	}

	session.StopSpinner()
	session.Printf("%s 客戶端創建: %s\n", icons.Info, clientID)
	return nil
}
