package ignitecmd

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/pkg/cliui/icons"
	"github.com/ignite-hq/cli/ignite/services/network"
	"github.com/ignite-hq/cli/ignite/services/network/networktypes"
)

var (
	campaignMainnetsAccSummaryHeader = []string{"Mainnet Account", "Shares"}
	campaignVestingAccSummaryHeader  = []string{"Vesting Account", "Total Shares", "Vesting", "End Time"}
)

// NewNetworkCampaignAccount 創建一個新的活動帳戶命令，其中包含其他一些
// 與活動帳戶相關的子命令。
func NewNetworkCampaignAccount() *cobra.Command {
	c := &cobra.Command{
		Use:   "account",
		Short: "處理活動帳戶",
	}
	c.AddCommand(
		newNetworkCampaignAccountList(),
	)
	return c
}

func newNetworkCampaignAccountList() *cobra.Command {
	c := &cobra.Command{
		Use:   "list [campaign-id]",
		Short: "顯示活動的所有主網和主網歸屬",
		Args:  cobra.ExactArgs(1),
		RunE:  newNetworkCampaignAccountListHandler,
	}
	return c
}

func newNetworkCampaignAccountListHandler(cmd *cobra.Command, args []string) error {
	session := cliui.New()
	defer session.Cleanup()

	nb, campaignID, err := networkChainLaunch(cmd, args, session)
	if err != nil {
		return err
	}
	n, err := nb.Network()
	if err != nil {
		return err
	}

	// 獲取所有活動帳戶
	mainnetAccs, vestingAccs, err := getAccounts(cmd.Context(), n, campaignID)
	if err != nil {
		return err
	}

	if len(mainnetAccs)+len(vestingAccs) == 0 {
		session.StopSpinner()
		return session.Printf("%s %s\n", icons.Info, "未找到活動中帳戶")
	}

	mainnetAccEntries := make([][]string, 0)
	for _, acc := range mainnetAccs {
		mainnetAccEntries = append(mainnetAccEntries, []string{acc.Address, acc.Shares.String()})
	}
	mainnetVestingAccEntries := make([][]string, 0)
	for _, acc := range vestingAccs {
		mainnetVestingAccEntries = append(mainnetVestingAccEntries, []string{
			acc.Address,
			acc.TotalShares.String(),
			acc.Vesting.String(),
			strconv.FormatInt(acc.EndTime, 10),
		})
	}

	session.StopSpinner()
	if len(mainnetAccEntries) > 0 {
		if err = session.PrintTable(campaignMainnetsAccSummaryHeader, mainnetAccEntries...); err != nil {
			return err
		}
	}
	if len(mainnetVestingAccEntries) > 0 {
		if err = session.PrintTable(campaignVestingAccSummaryHeader, mainnetVestingAccEntries...); err != nil {
			return err
		}
	}

	return nil
}

// getAccounts 獲取所有活動主網和歸屬賬戶。
func getAccounts(
	ctx context.Context,
	n network.Network,
	campaignID uint64,
) (
	[]networktypes.MainnetAccount,
	[]networktypes.MainnetVestingAccount,
	error,
) {
	// 開始服務組件。
	g, ctx := errgroup.WithContext(ctx)
	var (
		mainnetAccs []networktypes.MainnetAccount
		vestingAccs []networktypes.MainnetVestingAccount
		err         error
	)
	// 獲取所有競選主網賬戶
	g.Go(func() error {
		mainnetAccs, err = n.MainnetAccounts(ctx, campaignID)
		return err
	})

	// 獲取所有競選歸屬賬戶
	g.Go(func() error {
		vestingAccs, err = n.MainnetVestingAccounts(ctx, campaignID)
		return err
	})
	return mainnetAccs, vestingAccs, g.Wait()
}
