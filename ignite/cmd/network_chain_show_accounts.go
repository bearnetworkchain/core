package ignitecmd

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/pkg/cliui"
	"github.com/bearnetworkchain/core/ignite/pkg/cliui/icons"
	"github.com/bearnetworkchain/core/ignite/pkg/cosmosutil"
)

var (
	chainGenesisAccSummaryHeader = []string{"Genesis Account", "Coins"}
	chainVestingAccSummaryHeader = []string{"Vesting Account", "Total Balance", "Vesting", "EndTime"}
)

func newNetworkChainShowAccounts() *cobra.Command {
	c := &cobra.Command{
		Use:   "accounts [launch-id]",
		Short: "顯示鏈上的所有歸屬和創世賬戶",
		Args:  cobra.ExactArgs(1),
		RunE:  networkChainShowAccountsHandler,
	}

	c.Flags().AddFlagSet(flagSetSPNAccountPrefixes())

	return c
}

func networkChainShowAccountsHandler(cmd *cobra.Command, args []string) error {
	session := cliui.New()
	defer session.Cleanup()

	addressPrefix := getAddressPrefix(cmd)

	nb, launchID, err := networkChainLaunch(cmd, args, session)
	if err != nil {
		return err
	}
	n, err := nb.Network()
	if err != nil {
		return err
	}

	genesisAccs, err := n.GenesisAccounts(cmd.Context(), launchID)
	if err != nil {
		return err
	}
	vestingAccs, err := n.VestingAccounts(cmd.Context(), launchID)
	if err != nil {
		return err
	}
	if len(genesisAccs)+len(vestingAccs) == 0 {
		session.StopSpinner()
		return session.Printf("%s %s\n", icons.Info, "空鏈賬戶列表")
	}

	genesisAccEntries := make([][]string, 0)
	for _, acc := range genesisAccs {
		address, err := cosmosutil.ChangeAddressPrefix(acc.Address, addressPrefix)
		if err != nil {
			return err
		}

		genesisAccEntries = append(genesisAccEntries, []string{address, acc.Coins})
	}
	genesisVestingAccEntries := make([][]string, 0)
	for _, acc := range vestingAccs {
		address, err := cosmosutil.ChangeAddressPrefix(acc.Address, addressPrefix)
		if err != nil {
			return err
		}

		genesisVestingAccEntries = append(genesisVestingAccEntries, []string{
			address,
			acc.TotalBalance,
			acc.Vesting,
			strconv.FormatInt(acc.EndTime, 10),
		})
	}

	session.StopSpinner()
	if len(genesisAccEntries) > 0 {
		if err = session.PrintTable(chainGenesisAccSummaryHeader, genesisAccEntries...); err != nil {
			return err
		}
	}
	if len(genesisVestingAccEntries) > 0 {
		if err = session.PrintTable(chainVestingAccSummaryHeader, genesisVestingAccEntries...); err != nil {
			return err
		}
	}

	return nil
}
