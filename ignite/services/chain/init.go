package chain

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/imdario/mergo"

	"github.com/ignite-hq/cli/ignite/chainconfig"
	chaincmdrunner "github.com/ignite-hq/cli/ignite/pkg/chaincmd/runner"
	"github.com/ignite-hq/cli/ignite/pkg/confile"
)

const (
	moniker = "mynode"
)

// Init 初始化鏈並應用所有可選配置。
func (c *Chain) Init(ctx context.Context, initAccounts bool) error {
	conf, err := c.Config()
	if err != nil {
		return &CannotBuildAppError{err}
	}

	if err := c.InitChain(ctx); err != nil {
		return err
	}

	if initAccounts {
		return c.InitAccounts(ctx, conf)
	}
	return nil
}

// InitChain 初始化鏈。
func (c *Chain) InitChain(ctx context.Context) error {
	chainID, err := c.ID()
	if err != nil {
		return err
	}

	conf, err := c.Config()
	if err != nil {
		return err
	}

	// 從以前的“服務”中清除持久數據。
	home, err := c.Home()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(home); err != nil {
		return err
	}

	commands, err := c.Commands(ctx)
	if err != nil {
		return err
	}

	// 初始化節點。
	if err := commands.Init(ctx, moniker); err != nil {
		return err
	}

	// 將 Ignite CLI 的 config.yml 中的配置更改覆蓋到
	// 通過應用程序的 sdk 配置。

	if err := c.plugin.Configure(home, conf); err != nil {
		return err
	}

	//確保在 chain.New() 期間給出的鏈 ID 具有最高優先級.
	if conf.Genesis != nil {
		conf.Genesis["chain_id"] = chainID
	}

	// 初始化應用配置
	genesisPath, err := c.GenesisPath()
	if err != nil {
		return err
	}
	appTOMLPath, err := c.AppTOMLPath()
	if err != nil {
		return err
	}
	clientTOMLPath, err := c.ClientTOMLPath()
	if err != nil {
		return err
	}
	configTOMLPath, err := c.ConfigTOMLPath()
	if err != nil {
		return err
	}

	appconfigs := []struct {
		ec      confile.EncodingCreator
		path    string
		changes map[string]interface{}
	}{
		{confile.DefaultJSONEncodingCreator, genesisPath, conf.Genesis},
		{confile.DefaultTOMLEncodingCreator, appTOMLPath, conf.Init.App},
		{confile.DefaultTOMLEncodingCreator, clientTOMLPath, conf.Init.Client},
		{confile.DefaultTOMLEncodingCreator, configTOMLPath, conf.Init.Config},
	}

	for _, ac := range appconfigs {
		cf := confile.New(ac.ec, ac.path)
		var conf map[string]interface{}
		if err := cf.Load(&conf); err != nil {
			return err
		}
		if err := mergo.Merge(&conf, ac.changes, mergo.WithOverride); err != nil {
			return err
		}
		if err := cf.Save(conf); err != nil {
			return err
		}
	}

	return nil
}

//InitAccounts 初始化鏈賬戶並創建驗證者 gentxs
func (c *Chain) InitAccounts(ctx context.Context, conf chainconfig.Config) error {
	commands, err := c.Commands(ctx)
	if err != nil {
		return err
	}

	// 將賬戶從配置添加到創世
	for _, account := range conf.Accounts {
		var generatedAccount chaincmdrunner.Account
		accountAddress := account.Address

		// 如果帳戶沒有提供地址，我們會創建一個
		if accountAddress == "" {
			generatedAccount, err = commands.AddAccount(ctx, account.Name, account.Mnemonic, account.CoinType)
			if err != nil {
				return err
			}
			accountAddress = generatedAccount.Address
		}

		coins := strings.Join(account.Coins, ",")
		if err := commands.AddGenesisAccount(ctx, accountAddress, coins); err != nil {
			return err
		}

		if account.Address == "" {
			fmt.Fprintf(
				c.stdLog().out,
				"🙂 創建帳戶 %q 和地址 %q 和助記詞: %q\n",
				generatedAccount.Name,
				generatedAccount.Address,
				generatedAccount.Mnemonic,
			)
		} else {
			fmt.Fprintf(
				c.stdLog().out,
				"🙂 導入賬戶 %q 和地址: %q\n",
				account.Name,
				account.Address,
			)
		}
	}

	_, err = c.IssueGentx(ctx, Validator{
		Name:          conf.Validator.Name,
		StakingAmount: conf.Validator.Staked,
	})
	return err
}

// IssueGentx 從chain config中的validator信息生成一個gentx，並在chain genesis中導入
func (c Chain) IssueGentx(ctx context.Context, v Validator) (string, error) {
	commands, err := c.Commands(ctx)
	if err != nil {
		return "", err
	}

	// 從配置中的驗證器創建 gentx
	gentxPath, err := c.plugin.Gentx(ctx, commands, v)
	if err != nil {
		return "", err
	}

	// 將 gentx 導入創世紀
	return gentxPath, commands.CollectGentxs(ctx)
}

// IsInitialized 檢查鍊是否已初始化
// 通過檢查配置中是否存在 gentx 目錄來執行檢查
func (c *Chain) IsInitialized() (bool, error) {
	home, err := c.Home()
	if err != nil {
		return false, err
	}
	gentxDir := filepath.Join(home, "config", "gentx")

	if _, err := os.Stat(gentxDir); os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		// Return error on other error
		return false, err
	}

	return true, nil
}

type Validator struct {
	Name                    string
	Moniker                 string
	StakingAmount           string
	CommissionRate          string
	CommissionMaxRate       string
	CommissionMaxChangeRate string
	MinSelfDelegation       string
	GasPrices               string
	Details                 string
	Identity                string
	Website                 string
	SecurityContact         string
}

//Account 代錶鍊中的一個賬戶。
type Account struct {
	Name     string
	Address  string
	Mnemonic string `json:"mnemonic"`
	CoinType string
	Coins    string
}
