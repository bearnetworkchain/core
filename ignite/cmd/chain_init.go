package ignitecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/pkg/chaincmd"
	"github.com/bearnetworkchain/core/ignite/pkg/cliui/colors"
	"github.com/bearnetworkchain/core/ignite/services/chain"
)

func NewChainInit() *cobra.Command {
	c := &cobra.Command{
		Use:   "init",
		Short: "初始化熊網鏈",
		Args:  cobra.NoArgs,
		RunE:  chainInitHandler,
	}

	flagSetPath(c)
	flagSetClearCache(c)
	c.Flags().AddFlagSet(flagSetHome())

	return c
}

func chainInitHandler(cmd *cobra.Command, _ []string) error {
	chainOption := []chain.Option{
		chain.LogLevel(logLevel(cmd)),
		chain.KeyringBackend(chaincmd.KeyringBackendTest),
	}

	c, err := newChainWithHomeFlags(cmd, chainOption...)
	if err != nil {
		return err
	}

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	if _, err := c.Build(cmd.Context(), cacheStorage, ""); err != nil {
		return err
	}

	if err := c.Init(cmd.Context(), true); err != nil {
		return err
	}

	home, err := c.Home()
	if err != nil {
		return err
	}

	fmt.Printf("🗃  初始化。 簽出您的熊網鏈的主（數據）目錄: %s\n", colors.Info(home))

	return nil
}
