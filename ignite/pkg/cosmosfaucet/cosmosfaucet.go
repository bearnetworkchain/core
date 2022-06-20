// Package cosmosfaucet is a faucet to request tokens for sdk accounts.
package cosmosfaucet

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	chaincmdrunner "github.com/ignite-hq/cli/ignite/pkg/chaincmd/runner"
)

const (
	// DefaultAccountName 是從中轉移代幣的默認帳戶。
	DefaultAccountName = "faucet"

	// DefaultDenom 是分配的默認面額。
	DefaultDenom = "ubnkt"

	// DefaultAmount 指定轉入賬戶的默認金額
	// 在每個請求上。
	DefaultAmount = 10000000

	// DefaultMaxAmount 指定可以轉移到的最大金額
	// 隨時記賬。
	DefaultMaxAmount = 100000000

	// DefaultLimitRefreshWindow 指定最大數量限制之後的時間
	// 為帳戶刷新 [1 年]
	DefaultRefreshWindow = time.Hour * 24 * 365
)

// Faucet 代表水龍頭。
type Faucet struct {
	// runner 用於與區塊鏈二進制交互以傳輸令牌。
	runner chaincmdrunner.Runner

	// chainID 是個 chain id 水龍頭正在運行的鏈條。
	chainID string

	// accountName 從中轉移代幣。
	accountName string

	// accountMnemonic 是賬戶的助記詞。
	accountMnemonic string

	// coinType 註冊硬幣類型號碼熱錢包推導 (BIP-0044).
	coinType string

	// coins 保留可以由水龍頭分配的硬幣列表。
	coins sdk.Coins

	// coinsMax 是一個 denom-max 對。
	// 它擁有可以發送到單個帳戶的最大數量的硬幣。
	coinsMax map[string]uint64

	limitRefreshWindow time.Duration

	// openAPIData 保存用於服務 OpenAPI 頁面和規範的模板數據自定義。
	openAPIData openAPIData
}

// Option 配置水龍頭選項。
type Option func(*Faucet)

// Account 提供用於轉移代幣的賬戶信息。
// 如果沒有提供助記詞，則假定帳戶存在於密鑰環中。
func Account(name, mnemonic string, coinType string) Option {
	return func(f *Faucet) {
		f.accountName = name
		f.accountMnemonic = mnemonic
		f.coinType = coinType
	}
}

// Coin 將一個新的硬幣添加到硬幣列表中以通過水龍頭分發。
// 添加到列表中的第一個硬幣在轉移請求期間被視為默認硬幣。
//
// amount 是每個請求可以分配的硬幣數量。
// maxAmount 是可以發送到單個帳戶的最大硬幣數量。
// denom 是要通過水龍頭分配的硬幣的面額。
func Coin(amount, maxAmount uint64, denom string) Option {
	return func(f *Faucet) {
		f.coins = append(f.coins, sdk.NewCoin(denom, sdk.NewIntFromUint64(amount)))
		f.coinsMax[denom] = maxAmount
	}
}

// RefreshWindow 將刷新傳輸限制的持續時間添加到水龍頭
func RefreshWindow(refreshWindow time.Duration) Option {
	return func(f *Faucet) {
		f.limitRefreshWindow = refreshWindow
	}
}

// ChainID 添加 chain id 去水龍頭。 faucet 將在未提供時自動獲取。
func ChainID(id string) Option {
	return func(f *Faucet) {
		f.chainID = id
	}
}

// OpenAPI 配置如何提供 Open API 頁面和規範。
func OpenAPI(apiAddress string) Option {
	return func(f *Faucet) {
		f.openAPIData.APIAddress = apiAddress
	}
}

// New 使用 ccr（訪問和使用區塊鏈的 CLI）和給定選項創建一個新水龍頭。
func New(ctx context.Context, ccr chaincmdrunner.Runner, options ...Option) (Faucet, error) {
	f := Faucet{
		runner:      ccr,
		accountName: DefaultAccountName,
		coinsMax:    make(map[string]uint64),
		openAPIData: openAPIData{"Blockchain", "https://0.0.0.0:1317"},
	}

	for _, apply := range options {
		apply(&f)
	}

	if len(f.coins) == 0 {
		Coin(DefaultAmount, DefaultMaxAmount, DefaultDenom)(&f)
	}

	if f.limitRefreshWindow == 0 {
		RefreshWindow(DefaultRefreshWindow)(&f)
	}

	// 如果提供助記詞，則導入帳戶.
	if f.accountMnemonic != "" {
		_, err := f.runner.AddAccount(ctx, f.accountName, f.accountMnemonic, f.coinType)
		if err != nil && err != chaincmdrunner.ErrAccountAlreadyExists {
			return Faucet{}, err
		}
	}

	if f.chainID == "" {
		status, err := f.runner.Status(ctx)
		if err != nil {
			return Faucet{}, err
		}

		f.chainID = status.ChainID
		f.openAPIData.ChainID = status.ChainID
	}

	return f, nil
}
