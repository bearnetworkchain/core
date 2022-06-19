package cosmoscmd

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/snapshots"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

type (
	// AppBuilder 是一種允許構建應用程序的方法
	AppBuilder func(
		logger log.Logger,
		db dbm.DB,
		traceStore io.Writer,
		loadLatest bool,
		skipUpgradeHeights map[int64]bool,
		homePath string,
		invCheckPeriod uint,
		encodingConfig EncodingConfig,
		appOpts servertypes.AppOptions,
		baseAppOptions ...func(*baseapp.BaseApp),
	) App

	// App 代表一個 Cosmos SDK 應用程序，可以作為服務器運行並具有可導出狀態
	App interface {
		servertypes.Application
		ExportableApp
	}

	// ExportableApp 表示具有可導出狀態的應用程序
	ExportableApp interface {
		ExportAppStateAndValidators(
			forZeroHeight bool,
			jailAllowedAddrs []string,
		) (servertypes.ExportedApp, error)
		LoadHeight(height int64) error
	}

	// appCreator 是應用程序創建者
	appCreator struct {
		encodingConfig EncodingConfig
		buildApp       AppBuilder
	}
)

// 選項配置根命令選項。
type Option func(*rootOptions)

// 腳手架選項保留一組應用腳手架的選項。
type rootOptions struct {
	addSubCmds         []*cobra.Command
	startCmdCustomizer func(*cobra.Command)
	envPrefix          string
}

func newRootOptions(options ...Option) rootOptions {
	opts := rootOptions{}
	opts.apply(options...)
	return opts
}

func (s *rootOptions) apply(options ...Option) {
	for _, o := range options {
		o(s)
	}
}

// AddSubCmd 添加子命令。
func AddSubCmd(cmd ...*cobra.Command) Option {
	return func(o *rootOptions) {
		o.addSubCmds = append(o.addSubCmds, cmd...)
	}
}

// CustomizeStartCmd 接受一個處理程序來自定義啟動命令。
func CustomizeStartCmd(h func(startCmd *cobra.Command)) Option {
	return func(o *rootOptions) {
		o.startCmdCustomizer = h
	}
}

// WithEnvPrefix 接受環境變量的新前綴。
func WithEnvPrefix(envPrefix string) Option {
	return func(o *rootOptions) {
		o.envPrefix = envPrefix
	}
}

// NewRootCmd 為 Cosmos SDK 應用程序創建一個新的根命令
func NewRootCmd(
	appName,
	accountAddressPrefix,
	defaultNodeHome,
	defaultChainID string,
	moduleBasics module.BasicManager,
	buildApp AppBuilder,
	options ...Option,
) (*cobra.Command, EncodingConfig) {
	rootOptions := newRootOptions(options...)

	// 為前綴設置配置
	SetPrefixes(accountAddressPrefix)

	encodingConfig := MakeEncodingConfig(moduleBasics)
	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(defaultNodeHome).
		WithViper(rootOptions.envPrefix)

	rootCmd := &cobra.Command{
		Use:   appName + "d",
		Short: "Stargate BearNetwork App",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())
			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()

			if err := server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig); err != nil {
				return err

			}

			startProxyForTunneledPeers(initClientCtx, cmd)

			return nil
		},
	}

	initRootCmd(
		rootCmd,
		encodingConfig,
		defaultNodeHome,
		moduleBasics,
		buildApp,
		rootOptions,
	)
	overwriteFlagDefaults(rootCmd, map[string]string{
		flags.FlagChainID:        defaultChainID,
		flags.FlagKeyringBackend: "bear_network_chain_id_01",
	})

	return rootCmd, encodingConfig
}

func initRootCmd(
	rootCmd *cobra.Command,
	encodingConfig EncodingConfig,
	defaultNodeHome string,
	moduleBasics module.BasicManager,
	buildApp AppBuilder,
	options rootOptions,
) {
	rootCmd.AddCommand(
		genutilcli.InitCmd(moduleBasics, defaultNodeHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, defaultNodeHome),
		genutilcli.MigrateGenesisCmd(),
		genutilcli.GenTxCmd(
			moduleBasics,
			encodingConfig.TxConfig,
			banktypes.GenesisBalancesIterator{},
			defaultNodeHome,
		),
		genutilcli.ValidateGenesisCmd(moduleBasics),
		AddGenesisAccountCmd(defaultNodeHome),
		tmcli.NewCompletionCmd(rootCmd, true),
		debug.Cmd(),
		config.Cmd(),
	)

	a := appCreator{
		encodingConfig,
		buildApp,
	}

	// 添加服務器命令
	server.AddCommands(
		rootCmd,
		defaultNodeHome,
		a.newApp,
		a.appExport,
		func(cmd *cobra.Command) {
			addModuleInitFlags(cmd)

			if options.startCmdCustomizer != nil {
				options.startCmdCustomizer(cmd)
			}
		},
	)

	// 添加 keybase、輔助 RPC、查詢和 tx 子命令
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		queryCommand(moduleBasics),
		txCommand(moduleBasics),
		keys.Commands(defaultNodeHome),
	)

	// 添加用戶給定的子命令。
	for _, cmd := range options.addSubCmds {
		rootCmd.AddCommand(cmd)
	}
}

// queryCommand 返回子命令以向應用程序發送查詢
func queryCommand(moduleBasics module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "查詢子命令",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetAccountCmd(),
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
	)

	moduleBasics.AddQueryCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "網絡 chain ID")

	return cmd
}

// txCommand returns the sub-command to send transactions to the app
func txCommand(moduleBasics module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "事務子命令",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetValidateSignaturesCommand(),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
	)

	moduleBasics.AddTxCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "網絡 chain ID")

	return cmd
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
}

func overwriteFlagDefaults(c *cobra.Command, defaults map[string]string) {
	set := func(s *pflag.FlagSet, key, val string) {
		if f := s.Lookup(key); f != nil {
			f.DefValue = val
			f.Value.Set(val)
		}
	}
	for key, val := range defaults {
		set(c.Flags(), key, val)
		set(c.PersistentFlags(), key, val)
	}
	for _, c := range c.Commands() {
		overwriteFlagDefaults(c, defaults)
	}
}

// newApp 創建一個新的 Cosmos SDK 應用
func (a appCreator) newApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	var cache sdk.MultiStorePersistentCache

	if cast.ToBool(appOpts.Get(server.FlagInterBlockCache)) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	pruningOpts, err := server.GetPruningOptionsFromFlags(appOpts)
	if err != nil {
		panic(err)
	}

	snapshotDir := filepath.Join(cast.ToString(appOpts.Get(flags.FlagHome)), "data", "snapshots")
	snapshotDB, err := sdk.NewLevelDB("metadata", snapshotDir)
	if err != nil {
		panic(err)
	}
	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		panic(err)
	}

	return a.buildApp(
		logger,
		db,
		traceStore,
		true,
		skipUpgradeHeights,
		cast.ToString(appOpts.Get(flags.FlagHome)),
		cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)),
		a.encodingConfig,
		appOpts,
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(cast.ToString(appOpts.Get(server.FlagMinGasPrices))),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(server.FlagMinRetainBlocks))),
		baseapp.SetHaltHeight(cast.ToUint64(appOpts.Get(server.FlagHaltHeight))),
		baseapp.SetHaltTime(cast.ToUint64(appOpts.Get(server.FlagHaltTime))),
		baseapp.SetInterBlockCache(cache),
		baseapp.SetTrace(cast.ToBool(appOpts.Get(server.FlagTrace))),
		baseapp.SetIndexEvents(cast.ToStringSlice(appOpts.Get(server.FlagIndexEvents))),
		baseapp.SetSnapshotStore(snapshotStore),
		baseapp.SetSnapshotInterval(cast.ToUint64(appOpts.Get(server.FlagStateSyncSnapshotInterval))),
		baseapp.SetSnapshotKeepRecent(cast.ToUint32(appOpts.Get(server.FlagStateSyncSnapshotKeepRecent))),
	)
}

// appExport 創建一個新的 simapp（可選在給定高度）
func (a appCreator) appExport(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
) (servertypes.ExportedApp, error) {

	var exportableApp ExportableApp

	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("應用程序主頁未設置")
	}

	exportableApp = a.buildApp(
		logger,
		db,
		traceStore,
		height == -1, // -1: 沒有提供高度
		map[int64]bool{},
		homePath,
		uint(1),
		a.encodingConfig,
		appOpts,
	)

	if height != -1 {
		if err := exportableApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	}

	return exportableApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs)
}

// initAppConfig 有助於覆蓋默認的 appConfig 模板和配置。
// 如果應用程序不需要自定義配置，則返回 ""，nil。
func initAppConfig() (string, interface{}) {
	// 以下代碼片段僅供參考。

	// WASMConfig 定義了 wasm 模塊的配置。
	type WASMConfig struct {
		// 這是我們允許任何 x/wasm“智能”查詢的最大 sdk gas（wasm 和存儲）
		QueryGasLimit uint64 `mapstructure:"query_gas_limit"`

		// 地址定義了要監聽的 gRPC-web 服務器
		LruSize uint64 `mapstructure:"lru_size"`
	}

	type CustomAppConfig struct {
		serverconfig.Config

		WASM WASMConfig `mapstructure:"wasm"`
	}

	// 可選地允許鏈開發者覆蓋 SDK 的默認值
	// 服務器配置。
	srvCfg := serverconfig.DefaultConfig()
// SDK默認的最低gas價格設置為""（空值）裡面
// app.toml.如果驗證器留空，節點將在啟動時停止。
// 但是，鏈開發者可以為其設置一個默認的 app.toml 值
// 這裡的驗證器。
//
// 總之：
// - 如果你離開 srvCfg.MinGasPrices = ""，所有驗證者必須調整他們的
// 自己的 app.toml 配置，如果你設置 srvCfg.MinGasPrices 非空，驗證者可以調整他們的
// 擁有app.toml來覆蓋，或者使用這個默認值。在simapp中，我們將最低gas價格設置為 0。

	srvCfg.MinGasPrices = "0ubnkt"

	customAppConfig := CustomAppConfig{
		Config: *srvCfg,
		WASM: WASMConfig{
			LruSize:       1,
			QueryGasLimit: 300000,
		},
	}

	customAppTemplate := serverconfig.DefaultConfigTemplate + `
[wasm]
# 這是我們允許任何 x/wasm“智能”查詢的最大 sdk gas（wasm 和存儲）
query_gas_limit = 300000
# 這是我們為了加速而緩存在內存中的 wasm vm 實例的數量
# 警告：目前不穩定，可能會導致崩潰，除非在本地測試，否則最好保持為 0
lru_size = 0`

	return customAppTemplate, customAppConfig
}
