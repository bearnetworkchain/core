package ignitecmd

import (
	"fmt"

	"github.com/spf13/cobra"
	launchtypes "github.com/tendermint/spn/x/launch/types"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosutil"
	"github.com/ignite-hq/cli/ignite/services/network"
	"github.com/ignite-hq/cli/ignite/services/network/networktypes"
)

var requestSummaryHeader = []string{"ID", "Status", "Type", "Content"}

// NewNetworkRequestList 創建一個新的請求列表命令來列出
// 請求鏈
func NewNetworkRequestList() *cobra.Command {
	c := &cobra.Command{
		Use:   "list [launch-id]",
		Short: "列出所有待處理的請求",
		RunE:  networkRequestListHandler,
		Args:  cobra.ExactArgs(1),
	}

	c.Flags().AddFlagSet(flagSetSPNAccountPrefixes())

	return c
}

func networkRequestListHandler(cmd *cobra.Command, args []string) error {
	session := cliui.New()
	defer session.Cleanup()

	nb, err := newNetworkBuilder(cmd, CollectEvents(session.EventBus()))
	if err != nil {
		return err
	}

	addressPrefix := getAddressPrefix(cmd)

	// parse launch ID
	launchID, err := network.ParseID(args[0])
	if err != nil {
		return err
	}

	n, err := nb.Network()
	if err != nil {
		return err
	}

	requests, err := n.Requests(cmd.Context(), launchID)
	if err != nil {
		return err
	}

	session.StopSpinner()

	return renderRequestSummaries(requests, session, addressPrefix)
}

// renderRequestSummaries 寫入提供的輸出，匯總請求列表
func renderRequestSummaries(
	requests []networktypes.Request,
	session cliui.Session,
	addressPrefix string,
) error {
	requestEntries := make([][]string, 0)
	for _, request := range requests {
		var (
			id          = fmt.Sprintf("%d", request.RequestID)
			requestType = "Unknown"
			content     = ""
		)
		switch req := request.Content.Content.(type) {
		case *launchtypes.RequestContent_GenesisAccount:
			requestType = "添加創世紀賬戶"

			address, err := cosmosutil.ChangeAddressPrefix(
				req.GenesisAccount.Address,
				addressPrefix,
			)
			if err != nil {
				return err
			}

			content = fmt.Sprintf("%s, %s",
				address,
				req.GenesisAccount.Coins.String())
		case *launchtypes.RequestContent_GenesisValidator:
			requestType = "添加創世紀驗證器"
			peer, err := network.PeerAddress(req.GenesisValidator.Peer)
			if err != nil {
				return err
			}

			address, err := cosmosutil.ChangeAddressPrefix(
				req.GenesisValidator.Address,
				addressPrefix,
			)
			if err != nil {
				return err
			}

			content = fmt.Sprintf("%s, %s, %s",
				peer,
				address,
				req.GenesisValidator.SelfDelegation.String())
		case *launchtypes.RequestContent_VestingAccount:
			requestType = "添加歸屬賬戶"

			// parse vesting options
			var vestingCoins string
			dv := req.VestingAccount.VestingOptions.GetDelayedVesting()
			if dv == nil {
				vestingCoins = "無法識別的歸屬選項"
			} else {
				vestingCoins = fmt.Sprintf("%s (歸屬: %s)", dv.TotalBalance, dv.Vesting)
			}

			address, err := cosmosutil.ChangeAddressPrefix(
				req.VestingAccount.Address,
				addressPrefix,
			)
			if err != nil {
				return err
			}

			content = fmt.Sprintf("%s, %s",
				address,
				vestingCoins,
			)
		case *launchtypes.RequestContent_ValidatorRemoval:
			requestType = "移除驗證器"

			address, err := cosmosutil.ChangeAddressPrefix(
				req.ValidatorRemoval.ValAddress,
				addressPrefix,
			)
			if err != nil {
				return err
			}

			content = address
		case *launchtypes.RequestContent_AccountRemoval:
			requestType = "刪除帳戶"

			address, err := cosmosutil.ChangeAddressPrefix(
				req.AccountRemoval.Address,
				addressPrefix,
			)
			if err != nil {
				return err
			}

			content = address
		}

		requestEntries = append(requestEntries, []string{
			id,
			request.Status,
			requestType,
			content,
		})
	}
	return session.PrintTable(requestSummaryHeader, requestEntries...)
}
