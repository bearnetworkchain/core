package ignitecmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cosmosaccount"
)

// NewRelayer 返回一個新的中繼命令。
func NewRelayer() *cobra.Command {
	c := &cobra.Command{
		Use:     "relayer",
		Aliases: []string{"r"},
		Short:   "使用 IBC 協議連接區塊鏈",
	}

	c.AddCommand(
		NewRelayerConfigure(),
		NewRelayerConnect(),
	)

	return c
}

func handleRelayerAccountErr(err error) error {
	var accountErr *cosmosaccount.AccountDoesNotExistError
	if !errors.As(err, &accountErr) {
		return err
	}

	return errors.Wrap(accountErr, `確保通過“ignite account”命令創建或導入您的帳戶`)
}
