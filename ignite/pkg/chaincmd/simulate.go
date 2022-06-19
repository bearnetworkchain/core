package chaincmd

import (
	"path/filepath"
	"strconv"

	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"
	"github.com/ignite-hq/cli/ignite/pkg/gocmd"
)

const (
	optionSimappGenesis                = "-Genesis"
	optionSimappParams                 = "-Params"
	optionSimappExportParamsPath       = "-ExportParamsPath"
	optionSimappExportParamsHeight     = "-ExportParamsHeight"
	optionSimappExportStatePath        = "-ExportStatePath"
	optionSimappExportStatsPath        = "-ExportStatsPath"
	optionSimappSeed                   = "-Seed"
	optionSimappInitialBlockHeight     = "-InitialBlockHeight"
	optionSimappNumBlocks              = "-NumBlocks"
	optionSimappBlockSize              = "-BlockSize"
	optionSimappLean                   = "-Lean"
	optionSimappCommit                 = "-Commit"
	optionSimappSimulateEveryOperation = "-SimulateEveryOperation"
	optionSimappPrintAllInvariants     = "-PrintAllInvariants"
	optionSimappEnabled                = "-Enabled"
	optionSimappVerbose                = "-Verbose"
	optionSimappPeriod                 = "-Period"
	optionSimappGenesisTime            = "-GenesisTime"

	commandGoTest       = "test"
	optionGoBenchmem    = "-benchmem"
	optionGoSimappRun   = "-run=^$"
	optionGoSimappBench = "-bench=^BenchmarkSimulation"
)

// SimappOption 對於模擬命令
type SimappOption func([]string) []string

// SimappWithGenesis 為 simapp 命令提供創世紀選項
func SimappWithGenesis(genesis string) SimappOption {
	return func(command []string) []string {
		if len(genesis) > 0 {
			return append(command, optionSimappGenesis, genesis)
		}
		return command
	}
}

// SimappWithParams 為 simapp 命令提供 params 選項
func SimappWithParams(params string) SimappOption {
	return func(command []string) []string {
		if len(params) > 0 {
			return append(command, optionSimappParams, params)
		}
		return command
	}
}

// SimappWithExportParamsPath 為 simapp 命令提供輸出 Params路徑選項
func SimappWithExportParamsPath(exportParamsPath string) SimappOption {
	return func(command []string) []string {
		if len(exportParamsPath) > 0 {
			return append(command, optionSimappExportParamsPath, exportParamsPath)
		}
		return command
	}
}

// SimappWithExportParamsHeight 為 simapp 命令提供輸出 Params 高度選項
func SimappWithExportParamsHeight(exportParamsHeight int) SimappOption {
	return func(command []string) []string {
		if exportParamsHeight > 0 {
			return append(
				command,
				optionSimappExportParamsHeight,
				strconv.Itoa(exportParamsHeight),
			)
		}
		return command
	}
}

// SimappWithExportStatePath 提供導出狀態路徑 simapp 命令的選項
func SimappWithExportStatePath(exportStatePath string) SimappOption {
	return func(command []string) []string {
		if len(exportStatePath) > 0 {
			return append(command, optionSimappExportStatePath, exportStatePath)
		}
		return command
	}
}

// SimappWithExportStatsPath 為 simapp 命令提供導出統計路徑選項
func SimappWithExportStatsPath(exportStatsPath string) SimappOption {
	return func(command []string) []string {
		if len(exportStatsPath) > 0 {
			return append(command, optionSimappExportStatsPath, exportStatsPath)
		}
		return command
	}
}

// SimappWithSeed 為 simapp 命令提供種子選項
func SimappWithSeed(seed int64) SimappOption {
	return func(command []string) []string {
		return append(command, optionSimappSeed, strconv.FormatInt(seed, 10))
	}
}

// SimappWithInitialBlockHeight 為 simapp 命令提供初始塊高度選項
func SimappWithInitialBlockHeight(initialBlockHeight int) SimappOption {
	return func(command []string) []string {
		return append(command, optionSimappInitialBlockHeight, strconv.Itoa(initialBlockHeight))
	}
}

// SimappWithNumBlocks 為 simapp 命令提供塊數選項
func SimappWithNumBlocks(numBlocks int) SimappOption {
	return func(command []string) []string {
		return append(command, optionSimappNumBlocks, strconv.Itoa(numBlocks))
	}
}

// SimappWithBlockSize 為 simapp 命令提供塊大小選項
func SimappWithBlockSize(blockSize int) SimappOption {
	return func(command []string) []string {
		return append(command, optionSimappBlockSize, strconv.Itoa(blockSize))
	}
}

// SimappWithLean 為 simapp 命令提供精益選項
func SimappWithLean(lean bool) SimappOption {
	return func(command []string) []string {
		if lean {
			return append(command, optionSimappLean)
		}
		return command
	}
}

// SimappWithCommit 為 simapp 命令提供提交選項
func SimappWithCommit(commit bool) SimappOption {
	return func(command []string) []string {
		if commit {
			return append(command, optionSimappCommit)
		}
		return command
	}
}

// SimappWithSimulateEveryOperation 為 simapp 命令提供模擬每個操作選項
func SimappWithSimulateEveryOperation(simulateEveryOperation bool) SimappOption {
	return func(command []string) []string {
		if simulateEveryOperation {
			return append(command, optionSimappSimulateEveryOperation)
		}
		return command
	}
}

// SimappWithPrintAllInvariants 為 simapp 命令提供打印所有不變量選項
func SimappWithPrintAllInvariants(printAllInvariants bool) SimappOption {
	return func(command []string) []string {
		if printAllInvariants {
			return append(command, optionSimappPrintAllInvariants)
		}
		return command
	}
}

// SimappWithEnable 為 simapp 命令提供啟用選項
func SimappWithEnable(enable bool) SimappOption {
	return func(command []string) []string {
		if enable {
			return append(command, optionSimappEnabled)
		}
		return command
	}
}

// SimappWithVerbose 為 simapp 命令提供詳細選項
func SimappWithVerbose(verbose bool) SimappOption {
	return func(command []string) []string {
		if verbose {
			return append(command, optionSimappVerbose)
		}
		return command
	}
}

// SimappWithPeriod 為 simapp 命令提供句點選項
func SimappWithPeriod(period uint) SimappOption {
	return func(command []string) []string {
		return append(command, optionSimappPeriod, strconv.Itoa(int(period)))
	}
}

// SimappWithGenesisTime 為 simapp 命令提供創世時間選項
func SimappWithGenesisTime(genesisTime int64) SimappOption {
	return func(command []string) []string {
		return append(command, optionSimappGenesisTime, strconv.Itoa(int(genesisTime)))
	}
}

// SimulationCommand 返回用於 simapp 測試的 cli 命令
func SimulationCommand(appPath string, options ...SimappOption) step.Option {
	command := []string{
		commandGoTest,
		optionGoBenchmem,
		optionGoSimappRun,
		optionGoSimappBench,
		filepath.Join(appPath, "app"),
	}

	// 應用用戶提供的選項
	for _, applyOption := range options {
		command = applyOption(command)
	}
	return step.Exec(gocmd.Name(), command...)
}
