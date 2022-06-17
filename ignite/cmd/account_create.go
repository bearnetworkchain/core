package ignitecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/pkg/cosmosaccount"
)

func NewAccountCreate() *cobra.Command {
	c := &cobra.Command{
		Use:   "create [name]",
		Short: "創建一個新賬戶",
		Args:  cobra.ExactArgs(1),
		RunE:  accountCreateHandler,
	}

	c.Flags().AddFlagSet(flagSetKeyringBackend())

	return c
}

func accountCreateHandler(cmd *cobra.Command, args []string) error {
	name := args[0]

	ca, err := cosmosaccount.New(
		cosmosaccount.WithKeyringBackend(getKeyringBackend(cmd)),
	)
	if err != nil {
		return err
	}

	_, mnemonic, err := ca.Create(name)
	if err != nil {
		return err
	}

	fmt.Printf("帳戶 %q 已創建，請將您的助記符保存在一個秘密的地方:\n\n%s\n", name, mnemonic)
	return nil
}
