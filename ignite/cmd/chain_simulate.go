package ignitecmd

import (
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/services/chain"
)

const (
	flagSimappGenesis                = "genesis"
	flagSimappParams                 = "params"
	flagSimappExportParamsPath       = "exportParamsPath"
	flagSimappExportParamsHeight     = "exportParamsHeight"
	flagSimappExportStatePath        = "exportStatePath"
	flagSimappExportStatsPath        = "exportStatsPath"
	flagSimappSeed                   = "seed"
	flagSimappInitialBlockHeight     = "initialBlockHeight"
	flagSimappNumBlocks              = "numBlocks"
	flagSimappBlockSize              = "blockSize"
	flagSimappLean                   = "lean"
	flagSimappSimulateEveryOperation = "simulateEveryOperation"
	flagSimappPrintAllInvariants     = "printAllInvariants"
	flagSimappVerbose                = "verbose"
	flagSimappPeriod                 = "period"
	flagSimappGenesisTime            = "genesisTime"
)

//NewChainSimulate 創建一個新的模擬命令來運行區塊鏈模擬。
func NewChainSimulate() *cobra.Command {
	c := &cobra.Command{
		Use:   "simulate",
		Short: "對熊網區塊鏈進行模擬測試",
		Long:  "對熊網區塊鏈進行模擬測試。 它將每個模塊的許多隨機輸入消息發送到模擬節點並檢查不變量是否存在",
		Args:  cobra.NoArgs,
		RunE:  chainSimulationHandler,
	}
	simappFlags(c)
	return c
}

func chainSimulationHandler(cmd *cobra.Command, args []string) error {
	var (
		verbose, _     = cmd.Flags().GetBool(flagSimappVerbose)
		period, _      = cmd.Flags().GetUint(flagSimappPeriod)
		genesisTime, _ = cmd.Flags().GetInt64(flagSimappGenesisTime)
		config         = newConfigFromFlags(cmd)
		appPath        = flagGetPath(cmd)
	)
	// 用路徑創建鏈
	absPath, err := filepath.Abs(appPath)
	if err != nil {
		return err
	}
	c, err := chain.New(absPath)
	if err != nil {
		return err
	}

	config.ChainID, err = c.ID()
	if err != nil {
		return err
	}

	return c.Simulate(cmd.Context(),
		chain.SimappWithVerbose(verbose),
		chain.SimappWithPeriod(period),
		chain.SimappWithGenesisTime(genesisTime),
		chain.SimappWithConfig(config),
	)
}

// newConfigFromFlags 根據檢索到的標誌值創建模擬。
func newConfigFromFlags(cmd *cobra.Command) simulation.Config {
	var (
		genesis, _                = cmd.Flags().GetString(flagSimappGenesis)
		params, _                 = cmd.Flags().GetString(flagSimappParams)
		exportParamsPath, _       = cmd.Flags().GetString(flagSimappExportParamsPath)
		exportParamsHeight, _     = cmd.Flags().GetInt(flagSimappExportParamsHeight)
		exportStatePath, _        = cmd.Flags().GetString(flagSimappExportStatePath)
		exportStatsPath, _        = cmd.Flags().GetString(flagSimappExportStatsPath)
		seed, _                   = cmd.Flags().GetInt64(flagSimappSeed)
		initialBlockHeight, _     = cmd.Flags().GetInt(flagSimappInitialBlockHeight)
		numBlocks, _              = cmd.Flags().GetInt(flagSimappNumBlocks)
		blockSize, _              = cmd.Flags().GetInt(flagSimappBlockSize)
		lean, _                   = cmd.Flags().GetBool(flagSimappLean)
		simulateEveryOperation, _ = cmd.Flags().GetBool(flagSimappSimulateEveryOperation)
		printAllInvariants, _     = cmd.Flags().GetBool(flagSimappPrintAllInvariants)
	)
	return simulation.Config{
		Commit:             true,
		GenesisFile:        genesis,
		ParamsFile:         params,
		ExportParamsPath:   exportParamsPath,
		ExportParamsHeight: exportParamsHeight,
		ExportStatePath:    exportStatePath,
		ExportStatsPath:    exportStatsPath,
		Seed:               seed,
		InitialBlockHeight: initialBlockHeight,
		NumBlocks:          numBlocks,
		BlockSize:          blockSize,
		Lean:               lean,
		OnOperation:        simulateEveryOperation,
		AllInvariants:      printAllInvariants,
	}
}

func simappFlags(c *cobra.Command) {
	// config fields
	c.Flags().String(flagSimappGenesis, "", "自定義模擬創世紀文件； 不能與 params 文件一起使用")
	c.Flags().String(flagSimappParams, "", "覆蓋任何隨機參數的自定義模擬參數文件； 不能與創世紀一起使用")
	c.Flags().String(flagSimappExportParamsPath, "", "用於保存導出參數 JSON 的自定義文件路徑")
	c.Flags().Int(flagSimappExportParamsHeight, 0, "將隨機生成的參數導出到的高度")
	c.Flags().String(flagSimappExportStatePath, "", "用於保存導出的應用程序狀態 JSON 的自定義文件路徑")
	c.Flags().String(flagSimappExportStatsPath, "", "自定義文件路徑以保存導出的模擬統計信息 JSON")
	c.Flags().Int64(flagSimappSeed, 42, "模擬隨機種子")
	c.Flags().Int(flagSimappInitialBlockHeight, 1, "開始模擬的創世塊")
	c.Flags().Int(flagSimappNumBlocks, 200, "從初始塊高度模擬的新塊數")
	c.Flags().Int(flagSimappBlockSize, 30, "每個塊的操作")
	c.Flags().Bool(flagSimappLean, false, "精確模擬日誌輸出")
	c.Flags().Bool(flagSimappSimulateEveryOperation, false, "每次操作都運行緩慢的不變量")
	c.Flags().Bool(flagSimappPrintAllInvariants, false, "如果找到損壞的不變量，則打印所有不變量")

	// simulation flags
	c.Flags().BoolP(flagSimappVerbose, "v", false, "詳細日誌輸出")
	c.Flags().Uint(flagSimappPeriod, 0, "每個週期斷言只運行一次慢速不變量")
	c.Flags().Int64(flagSimappGenesisTime, 0, "覆蓋創世 UNIX 時間，而不是使用隨機的 UNIX 時間")
}
