package chain

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/pelletier/go-toml"

	"github.com/bearnetworkchain/core/ignite/chainconfig"
	"github.com/bearnetworkchain/core/ignite/pkg/chaincmd"
	chaincmdrunner "github.com/bearnetworkchain/core/ignite/pkg/chaincmd/runner"
	"github.com/bearnetworkchain/core/ignite/pkg/cosmosver"
	"github.com/bearnetworkchain/core/ignite/pkg/xurl"
)

type stargatePlugin struct {
	app App
}

func newStargatePlugin(app App) *stargatePlugin {
	return &stargatePlugin{
		app: app,
	}
}

func (p *stargatePlugin) Name() string {
	return "Stargate"
}

func (p *stargatePlugin) Gentx(ctx context.Context, runner chaincmdrunner.Runner, v Validator) (path string, err error) {
	return runner.Gentx(
		ctx,
		v.Name,
		v.StakingAmount,
		chaincmd.GentxWithMoniker(v.Moniker),
		chaincmd.GentxWithCommissionRate(v.CommissionRate),
		chaincmd.GentxWithCommissionMaxRate(v.CommissionMaxRate),
		chaincmd.GentxWithCommissionMaxChangeRate(v.CommissionMaxChangeRate),
		chaincmd.GentxWithMinSelfDelegation(v.MinSelfDelegation),
		chaincmd.GentxWithGasPrices(v.GasPrices),
		chaincmd.GentxWithDetails(v.Details),
		chaincmd.GentxWithIdentity(v.Identity),
		chaincmd.GentxWithWebsite(v.Website),
		chaincmd.GentxWithSecurityContact(v.SecurityContact),
	)
}

func (p *stargatePlugin) Configure(homePath string, conf chainconfig.Config) error {
	if err := p.appTOML(homePath, conf); err != nil {
		return err
	}
	if err := p.clientTOML(homePath); err != nil {
		return err
	}
	return p.configTOML(homePath, conf)
}

func (p *stargatePlugin) appTOML(homePath string, conf chainconfig.Config) error {
	// TODO 找到更好的方法，以便不刪除 toml.yml 中的註釋
	path := filepath.Join(homePath, "config/app.toml")
	config, err := toml.LoadFile(path)
	if err != nil {
		return err
	}

	apiAddr, err := xurl.TCP(conf.Host.API)
	if err != nil {
		return fmt.Errorf("api地址格式無效 %s: %w", conf.Host.API, err)
	}

	config.Set("api.enable", true)
	config.Set("api.enabled-unsafe-cors", true)
	config.Set("rpc.cors_allowed_origins", []string{"*"})
	config.Set("api.address", apiAddr)
	config.Set("grpc.address", conf.Host.GRPC)
	config.Set("grpc-web.address", conf.Host.GRPCWeb)

	staked, err := sdktypes.ParseCoinNormalized(conf.Validator.Staked)
	if err != nil {
		return err
	}
	gas := sdktypes.NewInt64Coin(staked.Denom, 0)
	config.Set("minimum-gas-prices", gas.String())

	file, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = config.WriteTo(file)
	return err
}

func (p *stargatePlugin) configTOML(homePath string, conf chainconfig.Config) error {
	// TODO 找到更好的方法，以便不刪除 toml.yml 中的註釋
	path := filepath.Join(homePath, "config/config.toml")
	config, err := toml.LoadFile(path)
	if err != nil {
		return err
	}

	rpcAddr, err := xurl.TCP(conf.Host.RPC)
	if err != nil {
		return fmt.Errorf("無效的 rpc 地址格式 %s: %w", conf.Host.RPC, err)
	}

	p2pAddr, err := xurl.TCP(conf.Host.P2P)
	if err != nil {
		return fmt.Errorf("無效的 p2p 地址格式 %s: %w", conf.Host.P2P, err)
	}

	config.Set("mode", "validator")
	config.Set("rpc.cors_allowed_origins", []string{"*"})
	config.Set("consensus.timeout_commit", "1s")
	config.Set("consensus.timeout_propose", "1s")
	config.Set("rpc.laddr", rpcAddr)
	config.Set("p2p.laddr", p2pAddr)
	config.Set("rpc.pprof_laddr", conf.Host.Prof)

	file, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = config.WriteTo(file)
	return err
}

func (p *stargatePlugin) clientTOML(homePath string) error {
	path := filepath.Join(homePath, "config/client.toml")
	config, err := toml.LoadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	config.Set("keyring-backend", "test")
	config.Set("broadcast-mode", "block")
	file, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = config.WriteTo(file)
	return err
}

func (p *stargatePlugin) Start(ctx context.Context, runner chaincmdrunner.Runner, conf chainconfig.Config) error {
	err := runner.Start(ctx,
		"--pruning",
		"nothing",
		"--grpc.address",
		conf.Host.GRPC,
	)
	return &CannotStartAppError{p.app.Name, err}
}

func (p *stargatePlugin) Home() string {
	return stargateHome(p.app)
}

func stargateHome(app App) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "."+app.Name)
}

func (p *stargatePlugin) Version() cosmosver.Family { return cosmosver.Stargate }

func (p *stargatePlugin) SupportsIBC() bool { return true }
