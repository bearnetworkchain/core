package networkchain

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	launchtypes "github.com/tendermint/spn/x/launch/types"

	"github.com/bearnetworkchain/core/ignite/pkg/cache"
	"github.com/bearnetworkchain/core/ignite/pkg/cosmosutil"
	"github.com/bearnetworkchain/core/ignite/pkg/events"
	"github.com/bearnetworkchain/core/ignite/services/network/networktypes"
)

// ResetGenesisTime 重置鏈創世時間
func (c Chain) ResetGenesisTime() error {
	// 設置鏈的創世時間
	genesisPath, err := c.GenesisPath()
	if err != nil {
		return errors.Wrap(err, "無法讀取區塊鏈的起源")
	}

	if err := cosmosutil.UpdateGenesis(
		genesisPath,
		cosmosutil.WithKeyValueTimestamp(cosmosutil.FieldGenesisTime, 0),
	); err != nil {
		return errors.Wrap(err, "創世時間無法設置")
	}
	return nil
}

// Prepare 準備從創世信息啟動的鏈
func (c Chain) Prepare(
	ctx context.Context,
	cacheStorage cache.Storage,
	gi networktypes.GenesisInformation,
	rewardsInfo networktypes.Reward,
	chainID string,
	lastBlockHeight,
	unbondingTime int64,
) error {
	// 鏈初始化
	genesisPath, err := c.chain.GenesisPath()
	if err != nil {
		return err
	}

	_, err = os.Stat(genesisPath)

	switch {
	case os.IsNotExist(err):
		// 如果不存在配置，則使用新的驗證器密鑰執行鏈的完整初始化
		if err = c.Init(ctx, cacheStorage); err != nil {
			return err
		}
	case err != nil:
		return err
	default:
		// 如果配置和驗證器密鑰已經存在，則構建鏈並初始化創世
		if _, err := c.Build(ctx, cacheStorage); err != nil {
			return err
		}

		if err := c.initGenesis(ctx); err != nil {
			return err
		}
	}

	if err := c.buildGenesis(
		ctx,
		gi,
		rewardsInfo,
		chainID,
		lastBlockHeight,
		unbondingTime,
	); err != nil {
		return err
	}

	cmd, err := c.chain.Commands(ctx)
	if err != nil {
		return err
	}

	// 確保創世紀具有有效的格式
	if err := cmd.ValidateGenesis(ctx); err != nil {
		return err
	}

	// 重置已保存的狀態，以防鏈之前已啟動
	if err := cmd.UnsafeReset(ctx); err != nil {
		return err
	}

	return nil
}

// buildGenesis 從啟動批准的請求中構建鏈的創世紀
func (c Chain) buildGenesis(
	ctx context.Context,
	gi networktypes.GenesisInformation,
	rewardsInfo networktypes.Reward,
	spnChainID string,
	lastBlockHeight,
	unbondingTime int64,
) error {
	c.ev.Send(events.New(events.StatusOngoing, "建立創世文件"))

	addressPrefix, err := c.detectPrefix(ctx)
	if err != nil {
		return errors.Wrap(err, "錯誤檢測鏈前綴")
	}

	// apply genesis information to the genesis
	if err := c.applyGenesisAccounts(ctx, gi.GenesisAccounts, addressPrefix); err != nil {
		return errors.Wrap(err, "將創世帳戶應用於創世時出錯")
	}
	if err := c.applyVestingAccounts(ctx, gi.VestingAccounts, addressPrefix); err != nil {
		return errors.Wrap(err, "將歸屬賬戶應用於創世記時出錯")
	}
	if err := c.applyGenesisValidators(ctx, gi.GenesisValidators); err != nil {
		return errors.Wrap(err, "將創世驗證器應用於創世時出錯")
	}

	genesisPath, err := c.chain.GenesisPath()
	if err != nil {
		return errors.Wrap(err, "無法讀取區塊鏈的創世文件")
	}

	// 更新創世紀
	if err := cosmosutil.UpdateGenesis(
		genesisPath,
		// 設置創世時間和鏈ID
		cosmosutil.WithKeyValue(cosmosutil.FieldChainID, c.id),
		cosmosutil.WithKeyValueTimestamp(cosmosutil.FieldGenesisTime, c.launchTime),
		// 設置網絡共識參數
		cosmosutil.WithKeyValue(cosmosutil.FieldConsumerChainID, spnChainID),
		cosmosutil.WithKeyValueInt(cosmosutil.FieldLastBlockHeight, lastBlockHeight),
		cosmosutil.WithKeyValue(cosmosutil.FieldConsensusTimestamp, rewardsInfo.ConsensusState.Timestamp),
		cosmosutil.WithKeyValue(cosmosutil.FieldConsensusNextValidatorsHash, rewardsInfo.ConsensusState.NextValidatorsHash),
		cosmosutil.WithKeyValue(cosmosutil.FieldConsensusRootHash, rewardsInfo.ConsensusState.Root.Hash),
		cosmosutil.WithKeyValueInt(cosmosutil.FieldConsumerUnbondingPeriod, unbondingTime),
		cosmosutil.WithKeyValueUint(cosmosutil.FieldConsumerRevisionHeight, rewardsInfo.RevisionHeight),
	); err != nil {
		return errors.Wrap(err, "創世時間無法設置")
	}

	c.ev.Send(events.New(events.StatusDone, "創世紀建成"))

	return nil
}

// applyGenesisAccounts 使用鏈 CLI 將創世賬戶添加到創世中
func (c Chain) applyGenesisAccounts(
	ctx context.Context,
	genesisAccs []networktypes.GenesisAccount,
	addressPrefix string,
) error {
	var err error

	cmd, err := c.chain.Commands(ctx)
	if err != nil {
		return err
	}

	for _, acc := range genesisAccs {
		// 將地址前綴更改為目標鏈前綴
		acc.Address, err = cosmosutil.ChangeAddressPrefix(acc.Address, addressPrefix)
		if err != nil {
			return err
		}

		// 調用 add genesis account CLI 命令
		err = cmd.AddGenesisAccount(ctx, acc.Address, acc.Coins)
		if err != nil {
			return err
		}
	}

	return nil
}

// applyVestingAccounts 使用鏈 CLI 將創世歸屬賬戶添加到創世中
func (c Chain) applyVestingAccounts(
	ctx context.Context,
	vestingAccs []networktypes.VestingAccount,
	addressPrefix string,
) error {
	cmd, err := c.chain.Commands(ctx)
	if err != nil {
		return err
	}

	for _, acc := range vestingAccs {
		acc.Address, err = cosmosutil.ChangeAddressPrefix(acc.Address, addressPrefix)
		if err != nil {
			return err
		}

		// 使用延遲歸屬選項調用 add genesis account CLI 命令
		err = cmd.AddVestingAccount(
			ctx,
			acc.Address,
			acc.TotalBalance,
			acc.Vesting,
			acc.EndTime,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// applyGenesisValidators 將驗證器 gentxs 收集到創世紀中並在配置中添加對等點
func (c Chain) applyGenesisValidators(ctx context.Context, genesisVals []networktypes.GenesisValidator) error {
	// 沒有驗證者
	if len(genesisVals) == 0 {
		return nil
	}

	// 重置 gentx 目錄
	gentxDir, err := c.chain.GentxsPath()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(gentxDir); err != nil {
		return err
	}
	if err := os.MkdirAll(gentxDir, 0700); err != nil {
		return err
	}

	// write gentxs
	for i, val := range genesisVals {
		gentxPath := filepath.Join(gentxDir, fmt.Sprintf("gentx%d.json", i))
		if err = ioutil.WriteFile(gentxPath, val.Gentx, 0666); err != nil {
			return err
		}
	}

	// gather gentxs
	cmd, err := c.chain.Commands(ctx)
	if err != nil {
		return err
	}
	if err := cmd.CollectGentxs(ctx); err != nil {
		return err
	}

	return c.updateConfigFromGenesisValidators(genesisVals)
}

// updateConfigFromGenesisValidators 將對等地址添加到鏈的 config.toml
func (c Chain) updateConfigFromGenesisValidators(genesisVals []networktypes.GenesisValidator) error {
	var (
		p2pAddresses    []string
		tunnelAddresses []TunneledPeer
	)
	for i, val := range genesisVals {
		if !cosmosutil.VerifyPeerFormat(val.Peer) {
			return errors.Errorf("無效對等: %s", val.Peer.Id)
		}
		switch conn := val.Peer.Connection.(type) {
		case *launchtypes.Peer_TcpAddress:
			p2pAddresses = append(p2pAddresses, fmt.Sprintf("%s@%s", val.Peer.Id, conn.TcpAddress))
		case *launchtypes.Peer_HttpTunnel:
			tunneledPeer := TunneledPeer{
				Name:      conn.HttpTunnel.Name,
				Address:   conn.HttpTunnel.Address,
				NodeID:    val.Peer.Id,
				LocalPort: strconv.Itoa(i + 22000),
			}
			tunnelAddresses = append(tunnelAddresses, tunneledPeer)
			p2pAddresses = append(p2pAddresses, fmt.Sprintf("%s@127.0.0.1:%s", tunneledPeer.NodeID, tunneledPeer.LocalPort))
		default:
			return fmt.Errorf("無效的對等類型")
		}
	}

	if len(p2pAddresses) > 0 {
		// 設置持久的對等點
		configPath, err := c.chain.ConfigTOMLPath()
		if err != nil {
			return err
		}
		configToml, err := toml.LoadFile(configPath)
		if err != nil {
			return err
		}
		configToml.Set("p2p.persistent_peers", strings.Join(p2pAddresses, ","))
		if err != nil {
			return err
		}

		// 如果有隧道對等點，它們將通過 localhost 與隧道客戶端連接，
		// 所以我們需要允許少數節點具有相同的 ip
		if len(tunnelAddresses) > 0 {
			configToml.Set("p2p.allow_duplicate_ip", true)
		}

		// 保存 config.toml 文件
		configTomlFile, err := os.OpenFile(configPath, os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer configTomlFile.Close()

		if _, err = configToml.WriteTo(configTomlFile); err != nil {
			return err
		}
	}

	if len(tunnelAddresses) > 0 {
		tunneledPeersConfigPath, err := c.SPNConfigPath()
		if err != nil {
			return err
		}

		if err = SetSPNConfig(Config{
			TunneledPeers: tunnelAddresses,
		}, tunneledPeersConfigPath); err != nil {
			return err
		}
	}
	return nil

}
