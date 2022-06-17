package chain

import (
	"context"
	"fmt"
	"os"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	chaincmdrunner "github.com/bearnetworkchain/core/ignite/pkg/chaincmd/runner"
	"github.com/bearnetworkchain/core/ignite/pkg/cosmosfaucet"
	"github.com/bearnetworkchain/core/ignite/pkg/xurl"
)

var (
	// ErrFaucetIsNotEnabled在 config.yml 中未啟用水龍頭時返回。
	ErrFaucetIsNotEnabled = errors.New("config.yml 中未啟用水龍頭")

	// ErrFaucetAccountDoesNotExist 當 config.yml 中指定的水龍頭賬戶不存在時返回。
	ErrFaucetAccountDoesNotExist = errors.New("指定的帳戶（faucet.name）不存在")
)

var (
	envAPIAddress = os.Getenv("API_ADDRESS")
)

// Faucet 返回鏈的水龍頭，如果水龍頭返回錯誤
// 配置錯誤或根本未配置（未啟用）。
func (c *Chain) Faucet(ctx context.Context) (cosmosfaucet.Faucet, error) {
	id, err := c.ID()
	if err != nil {
		return cosmosfaucet.Faucet{}, err
	}

	conf, err := c.Config()
	if err != nil {
		return cosmosfaucet.Faucet{}, err
	}

	commands, err := c.Commands(ctx)
	if err != nil {
		return cosmosfaucet.Faucet{}, err
	}

	// 驗證 config.yml 中的水龍頭初始化是否正確。
	if conf.Faucet.Name == nil {
		return cosmosfaucet.Faucet{}, ErrFaucetIsNotEnabled
	}

	if _, err := commands.ShowAccount(ctx, *conf.Faucet.Name); err != nil {
		if err == chaincmdrunner.ErrAccountDoesNotExist {
			return cosmosfaucet.Faucet{}, ErrFaucetAccountDoesNotExist
		}
		return cosmosfaucet.Faucet{}, err
	}

	// 構建水龍頭選項。
	apiAddress := conf.Host.API
	if envAPIAddress != "" {
		apiAddress = envAPIAddress
	}

	apiAddress, err = xurl.HTTP(apiAddress)
	if err != nil {
		return cosmosfaucet.Faucet{}, fmt.Errorf("無效的主機 api 地址格式: %w", err)
	}

	faucetOptions := []cosmosfaucet.Option{
		cosmosfaucet.Account(*conf.Faucet.Name, "", ""),
		cosmosfaucet.ChainID(id),
		cosmosfaucet.OpenAPI(apiAddress),
	}

	// 解析硬幣以作為硬幣傳遞給水龍頭。
	for _, coin := range conf.Faucet.Coins {
		parsedCoin, err := sdk.ParseCoinNormalized(coin)
		if err != nil {
			return cosmosfaucet.Faucet{}, fmt.Errorf("%s: %s", err, coin)
		}

		var amountMax uint64

		// 找出這枚硬幣的最大金額。
		for _, coinMax := range conf.Faucet.CoinsMax {
			parsedMax, err := sdk.ParseCoinNormalized(coinMax)
			if err != nil {
				return cosmosfaucet.Faucet{}, fmt.Errorf("%s: %s", err, coin)
			}
			if parsedMax.Denom == parsedCoin.Denom {
				amountMax = parsedMax.Amount.Uint64()
				break
			}
		}

		faucetOptions = append(faucetOptions, cosmosfaucet.Coin(parsedCoin.Amount.Uint64(), amountMax, parsedCoin.Denom))
	}

	if conf.Faucet.RateLimitWindow != "" {
		rateLimitWindow, err := time.ParseDuration(conf.Faucet.RateLimitWindow)
		if err != nil {
			return cosmosfaucet.Faucet{}, fmt.Errorf("%s: %s", err, conf.Faucet.RateLimitWindow)
		}

		faucetOptions = append(faucetOptions, cosmosfaucet.RefreshWindow(rateLimitWindow))
	}

	// 使用選項初始化水龍頭並返回。
	return cosmosfaucet.New(ctx, commands, faucetOptions...)
}
