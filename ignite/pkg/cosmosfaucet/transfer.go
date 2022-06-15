package cosmosfaucet

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	chaincmdrunner "github.com/ignite-hq/cli/ignite/pkg/chaincmd/runner"
)

// transferMutex 是一個互斥鎖，用於將傳輸請求保留在隊列中，因此檢查帳戶餘額和發送令牌是熊網幣的
var transferMutex = &sync.Mutex{}

// TotalTransferredAmount 返回從水龍頭賬戶轉賬到toAccountAddress的總金額.
func (f Faucet) TotalTransferredAmount(ctx context.Context, toAccountAddress, denom string) (totalAmount uint64, err error) {
	fromAccount, err := f.runner.ShowAccount(ctx, f.accountName)
	if err != nil {
		return 0, err
	}

	events, err := f.runner.QueryTxEvents(ctx,
		chaincmdrunner.NewEventSelector("message", "sender", fromAccount.Address),
		chaincmdrunner.NewEventSelector("transfer", "recipient", toAccountAddress))
	if err != nil {
		return 0, err
	}

	for _, event := range events {
		if event.Type == "transfer" {
			for _, attr := range event.Attributes {
				if attr.Key == "amount" {
					coins, err := sdk.ParseCoinsNormalized(attr.Value)
					if err != nil {
						return 0, err
					}

					amount := coins.AmountOf(denom).Uint64()

					if amount > 0 && time.Since(event.Time) < f.limitRefreshWindow {
						totalAmount += amount
					}
				}
			}
		}
	}

	return totalAmount, nil
}

// Transfer 將代幣數量從水龍頭賬戶轉移到toAccountAddress。
func (f *Faucet) Transfer(ctx context.Context, toAccountAddress string, coins sdk.Coins) error {
	transferMutex.Lock()
	defer transferMutex.Unlock()

	var coinsStr []string

	// 檢查每個硬幣，尚未達到最大轉賬金額
	for _, c := range coins {
		totalSent, err := f.TotalTransferredAmount(ctx, toAccountAddress, c.Denom)
		if err != nil {
			return err
		}

		if f.coinsMax[c.Denom] != 0 {
			if totalSent >= f.coinsMax[c.Denom] {
				return fmt.Errorf(
					"帳戶已達到最大值。允許金額 (%d) 為了 %q denom",
					f.coinsMax[c.Denom],
					c.Denom,
				)
			}

			if (totalSent + c.Amount.Uint64()) > f.coinsMax[c.Denom] {
				return fmt.Errorf(
					`要求少一些 %q denom. 帳戶已達到上限 (%d) 那個水龍頭可以忍受`,
					c.Denom,
					f.coinsMax[c.Denom],
				)
			}
		}

		coinsStr = append(coinsStr, c.String())
	}

	// 對所有硬幣進行轉賬
	fromAccount, err := f.runner.ShowAccount(ctx, f.accountName)
	if err != nil {
		return err
	}
	txHash, err := f.runner.BankSend(ctx, fromAccount.Address, toAccountAddress, strings.Join(coinsStr, ","))
	if err != nil {
		return err
	}

	// 等待發送 tx 被確認
	return f.runner.WaitTx(ctx, txHash, time.Second, 30)
}
