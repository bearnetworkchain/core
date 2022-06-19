package chain

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types/simulation"
)

type simappOptions struct {
	enabled     bool
	verbose     bool
	config      simulation.Config
	period      uint
	genesisTime int64
}

func newSimappOptions() simappOptions {
	return simappOptions{
		config: simulation.Config{
			Commit: true,
		},
		enabled:     true,
		verbose:     false,
		period:      0,
		genesisTime: 0,
	}
}

//SimappOption 為 simapp 命令提供選項
type SimappOption func(*simappOptions)

// SimappWithVerbose 啟用詳細模式
func SimappWithVerbose(verbose bool) SimappOption {
	return func(c *simappOptions) {
		c.verbose = verbose
	}
}

// SimappWithPeriod 允許每個週期斷言只運行一次慢速不變量
func SimappWithPeriod(period uint) SimappOption {
	return func(c *simappOptions) {
		c.period = period
	}
}

// SimappWithGenesisTime 允許覆蓋創世 UNIX 時間，而不是使用隨機 UNIX 時間
func SimappWithGenesisTime(genesisTime int64) SimappOption {
	return func(c *simappOptions) {
		c.genesisTime = genesisTime
	}
}

//SimappWithConfig 允許添加模擬配置
func SimappWithConfig(config simulation.Config) SimappOption {
	return func(c *simappOptions) {
		c.config = config
	}
}

func (c *Chain) Simulate(ctx context.Context, options ...SimappOption) error {
	simappOptions := newSimappOptions()

	// 應用選項
	for _, apply := range options {
		apply(&simappOptions)
	}

	commands, err := c.Commands(ctx)
	if err != nil {
		return err
	}
	return commands.Simulation(ctx,
		c.app.Path,
		simappOptions.enabled,
		simappOptions.verbose,
		simappOptions.config,
		simappOptions.period,
		simappOptions.genesisTime,
	)
}
