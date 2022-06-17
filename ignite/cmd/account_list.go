package ignitecmd

import (
	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/pkg/cosmosaccount"
)

func NewAccountList() *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "顯示所有帳戶的列表",
		RunE:  accountListHandler,
	}

	c.Flags().AddFlagSet(flagSetKeyringBackend())
	c.Flags().AddFlagSet(flagSetAccountPrefixes())

	return c
}

func accountListHandler(cmd *cobra.Command, args []string) error {
	ca, err := cosmosaccount.New(
		cosmosaccount.WithKeyringBackend(getKeyringBackend(cmd)),
	)
	if err != nil {
		return err
	}

	accounts, err := ca.List()
	if err != nil {
		return err
	}

	return printAccounts(cmd, accounts...)
}
