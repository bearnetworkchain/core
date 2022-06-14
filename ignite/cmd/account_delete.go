package ignitecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cosmosaccount"
)

func NewAccountDelete() *cobra.Command {
	c := &cobra.Command{
		Use:   "delete [name]",
		Short: "按名稱刪除帳戶",
		Args:  cobra.ExactArgs(1),
		RunE:  accountDeleteHandler,
	}

	c.Flags().AddFlagSet(flagSetKeyringBackend())

	return c
}

func accountDeleteHandler(cmd *cobra.Command, args []string) error {
	name := args[0]

	ca, err := cosmosaccount.New(
		cosmosaccount.WithKeyringBackend(getKeyringBackend(cmd)),
	)
	if err != nil {
		return err
	}

	if err := ca.DeleteByName(name); err != nil {
		return err
	}

	fmt.Printf("帳戶 %s 已刪除.\n", name)
	return nil
}
