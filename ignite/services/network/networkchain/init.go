package networkchain

import (
	"context"
	"fmt"
	"os"

	"github.com/ignite-hq/cli/ignite/pkg/cache"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosutil"
	"github.com/ignite-hq/cli/ignite/pkg/events"
)

// Init 通過構建二進製文件並運行 init 命令來初始化區塊鏈
// 創建鏈的初始創世，並設置驗證者密鑰
func (c *Chain) Init(ctx context.Context, cacheStorage cache.Storage) error {
	chainHome, err := c.chain.Home()
	if err != nil {
		return err
	}

	//清理應用程序的主目錄（如果存在）。
	if err = os.RemoveAll(chainHome); err != nil {
		return err
	}

	//構建鏈並使用新的驗證器密鑰對其進行初始化
	if _, err := c.Build(ctx, cacheStorage); err != nil {
		return err
	}

	c.ev.Send(events.New(events.StatusOngoing, "Initializing the blockchain"))

	if err = c.chain.Init(ctx, false); err != nil {
		return err
	}

	c.ev.Send(events.New(events.StatusDone, "Blockchain initialized"))

	//初始化並驗證創世紀
	if err = c.initGenesis(ctx); err != nil {
		return err
	}

	c.isInitialized = true

	return nil
}

//initGenesis 根據初始創世類型（默認，url，...）創建創世的初始創世
func (c *Chain) initGenesis(ctx context.Context) error {
	c.ev.Send(events.New(events.StatusOngoing, "計算創世紀"))

	genesisPath, err := c.chain.GenesisPath()
	if err != nil {
		return err
	}

	// remove existing genesis
	if err := os.RemoveAll(genesisPath); err != nil {
		return err
	}

	// 如果區塊鏈有創世 URL，則從該 URL 獲取初始創世
	// 否則，使用默認創世，不需要任何操作，因為默認創世是從 init 命令生成的
	if c.genesisURL != "" {
		genesis, hash, err := cosmosutil.GenesisAndHashFromURL(ctx, c.genesisURL)
		if err != nil {
			return err
		}

		//如果區塊鏈已經初始化且沒有創世哈希，我們將獲取的哈希分配給它
		//否則我們用現有的哈希檢查創世完整性
		if c.genesisHash == "" {
			c.genesisHash = hash
		} else if hash != c.genesisHash {
			return fmt.Errorf("創世於 URL %s 是無效的。預期哈希 %s, 實際哈希 %s", c.genesisURL, c.genesisHash, hash)
		}

		// 用獲取的創世紀,替換默認的創世紀
		if err := os.WriteFile(genesisPath, genesis, 0644); err != nil {
			return err
		}
	} else {
		// 使用默認創世紀，使用 init CLI 命令生成它
		cmd, err := c.chain.Commands(ctx)
		if err != nil {
			return err
		}

		// TODO: use validator moniker https://github.com/ignite-hq/cli/issues/1834
		if err := cmd.Init(ctx, "moniker"); err != nil {
			return err
		}

	}

	// check the genesis is valid
	if err := c.checkGenesis(ctx); err != nil {
		return err
	}

	c.ev.Send(events.New(events.StatusDone, "創世紀初始化"))
	return nil
}

// checkGenesis 檢查存儲的創世紀是否有效
func (c *Chain) checkGenesis(ctx context.Context) error {
	//使用 validate-genesis 命令對鏈執行靜態分析。
	chainCmd, err := c.chain.Commands(ctx)
	if err != nil {
		return err
	}

	return chainCmd.ValidateGenesis(ctx)

// TODO: 使用 validate-genesis 對 genesis 進行靜態分析不會檢查 genesis 的完全有效性
// 示例：不檢查 gentxs 格式
// 要對創世進行完整的有效性檢查，我們必須嘗試使用示例賬戶啟動鏈
}
