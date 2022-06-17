package ignitecmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/pkg/cliui/clispinner"
	"github.com/bearnetworkchain/core/ignite/pkg/placeholder"
	"github.com/bearnetworkchain/core/ignite/services/scaffolder"
)

// NewScaffoldBandchain 在模塊中創建一個新的 BandChain 預言機
func NewScaffoldBandchain() *cobra.Command {
	c := &cobra.Command{
		Use:   "band [queryName] --module [moduleName]",
		Short: "搭建 IBC BandChain 查詢預言機以請求實時數據",
		Long:  "在特定的啟用 IBC 的 Cosmos SDK 模塊中搭建 IBC BandChain 查詢預言機以從 BandChain 腳本請求實時數據e",
		Args:  cobra.MinimumNArgs(1),
		RunE:  createBandchainHandler,
	}

	flagSetPath(c)
	flagSetClearCache(c)
	c.Flags().String(flagModule, "", "IBC 模塊將數據包添加到")
	c.Flags().String(flagSigner, "", "消息簽名者的標籤（默認值：creator)")

	return c
}

func createBandchainHandler(cmd *cobra.Command, args []string) error {
	var (
		oracle  = args[0]
		appPath = flagGetPath(cmd)
		signer  = flagGetSigner(cmd)
	)

	s := clispinner.New().SetText("安裝腳手架...")
	defer s.Stop()

	module, err := cmd.Flags().GetString(flagModule)
	if err != nil {
		return err
	}
	if module == "" {
		return errors.New("請指定一個模塊來創建 BandChain 預言機: --module <模塊名稱>")
	}

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	var options []scaffolder.OracleOption
	if signer != "" {
		options = append(options, scaffolder.OracleWithSigner(signer))
	}

	sc, err := newApp(appPath)
	if err != nil {
		return err
	}

	sm, err := sc.AddOracle(cacheStorage, placeholder.New(), module, oracle, options...)
	if err != nil {
		return err
	}

	s.Stop()

	modificationsStr, err := sourceModificationToString(sm)
	if err != nil {
		return err
	}

	fmt.Println(modificationsStr)

	fmt.Printf(`
🎉 創建了一個Band預言機查詢 "%[1]v".

注意：BandChain 模塊使用版本“bandchain-1”。
確保相應地更新 keys.go 文件。

// x/%[2]v/types/keys.go
常量版本 = "bandchain-1"

`, oracle, module)

	return nil
}
