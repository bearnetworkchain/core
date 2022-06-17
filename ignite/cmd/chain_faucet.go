package ignitecmd

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/pkg/chaincmd"
	"github.com/bearnetworkchain/core/ignite/services/chain"
)

// NewChainFaucet 創建一個新的水龍頭命令來向賬戶發送硬幣。
func NewChainFaucet() *cobra.Command {
	c := &cobra.Command{
		Use:   "faucet [address] [coin<,...>]",
		Short: "將硬幣發送到帳戶",
		Args:  cobra.ExactArgs(2),
		RunE:  chainFaucetHandler,
	}

	flagSetPath(c)
	c.Flags().AddFlagSet(flagSetHome())
	c.Flags().BoolP("verbose", "v", false, "Verbose output")

	return c
}

func chainFaucetHandler(cmd *cobra.Command, args []string) error {
	var (
		toAddress = args[0]
		coins     = args[1]
	)

	chainOption := []chain.Option{
		chain.LogLevel(logLevel(cmd)),
		chain.KeyringBackend(chaincmd.KeyringBackendTest),
	}

	c, err := newChainWithHomeFlags(cmd, chainOption...)
	if err != nil {
		return err
	}

	faucet, err := c.Faucet(cmd.Context())
	if err != nil {
		return err
	}

	// 解析提供的硬幣
	parsedCoins, err := sdk.ParseCoinsNormalized(coins)
	if err != nil {
		return err
	}

	// 從水龍頭執行轉移
	if err := faucet.Transfer(cmd.Context(), toAddress, parsedCoins); err != nil {
		return err
	}

	fmt.Println("📨 發送的硬幣.")
	return nil
}
