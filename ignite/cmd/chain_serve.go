package ignitecmd

import (
	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/services/chain"
)

const (
	flagForceReset = "force-reset"
	flagResetOnce  = "reset-once"
	flagConfig     = "config"
)

// NewChainServe 創建一個新的服務命令來服務於區塊鏈。
func NewChainServe() *cobra.Command {
	c := &cobra.Command{
		Use:   "serve",
		Short: "在開發中啟動一個區塊鏈節點",
		Long:  "啟動具有自動重新加載功能的區塊鏈節點",
		Args:  cobra.NoArgs,
		RunE:  chainServeHandler,
	}

	flagSetPath(c)
	flagSetClearCache(c)
	c.Flags().AddFlagSet(flagSetHome())
	c.Flags().AddFlagSet(flagSetProto3rdParty(""))
	c.Flags().BoolP("verbose", "v", false, "詳細輸出")
	c.Flags().BoolP(flagForceReset, "f", false, "在啟動和每次項目源更改時,強制重置應用程序狀態")
	c.Flags().BoolP(flagResetOnce, "r", false, "首次啟動時重置應用程序狀態")
	c.Flags().StringP(flagConfig, "c", "", "熊網鏈配置文件 (default: ./config.yml)")

	return c
}

func chainServeHandler(cmd *cobra.Command, args []string) error {
	chainOption := []chain.Option{
		chain.LogLevel(logLevel(cmd)),
	}

	if flagGetProto3rdParty(cmd) {
		chainOption = append(chainOption, chain.EnableThirdPartyModuleCodegen())
	}

	// 檢查是否定義了自定義配置
	config, err := cmd.Flags().GetString(flagConfig)
	if err != nil {
		return err
	}
	if config != "" {
		chainOption = append(chainOption, chain.ConfigFile(config))
	}

	// 創建鏈
	c, err := newChainWithHomeFlags(cmd, chainOption...)
	if err != nil {
		return err
	}

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	// 服務於鏈條
	var serveOptions []chain.ServeOption
	forceUpdate, err := cmd.Flags().GetBool(flagForceReset)
	if err != nil {
		return err
	}
	if forceUpdate {
		serveOptions = append(serveOptions, chain.ServeForceReset())
	}
	resetOnce, err := cmd.Flags().GetBool(flagResetOnce)
	if err != nil {
		return err
	}
	if resetOnce {
		serveOptions = append(serveOptions, chain.ServeResetOnce())
	}

	return c.Serve(cmd.Context(), cacheStorage, serveOptions...)
}
