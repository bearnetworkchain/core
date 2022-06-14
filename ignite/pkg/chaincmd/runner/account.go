package chaincmdrunner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"
)

var (
	// ErrAccountAlreadyExists 在嘗試導入已存在的帳戶時返回。
	ErrAccountAlreadyExists = errors.New("賬戶已存在")

	// 賬戶未退出時返回 ErrAccountDoesNotExist.
	ErrAccountDoesNotExist = errors.New("賬戶無法退出")
)

// Account 代表一個用戶帳戶.
type Account struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Mnemonic string `json:"mnemonic,omitempty"`
}

// AddAccount 在提供助記詞時創建一個新帳戶或導入一個帳戶。
// 如果操作失敗或具有提供的名稱的帳戶返回錯誤
// 已經存在.
func (r Runner) AddAccount(ctx context.Context, name, mnemonic, coinType string) (Account, error) {
	if err := r.CheckAccountExist(ctx, name); err != nil {
		return Account{}, err
	}
	b := newBuffer()

	account := Account{
		Name:     name,
		Mnemonic: mnemonic,
	}

	// 提供助記詞時導入賬戶，否則創建新賬戶。
	if mnemonic != "" {
		input := &bytes.Buffer{}
		fmt.Fprintln(input, mnemonic)

		if r.chainCmd.KeyringPassword() != "" {
			fmt.Fprintln(input, r.chainCmd.KeyringPassword())
			fmt.Fprintln(input, r.chainCmd.KeyringPassword())
		}

		if err := r.run(
			ctx,
			runOptions{},
			r.chainCmd.RecoverKeyCommand(name, coinType),
			step.Write(input.Bytes()),
		); err != nil {
			return Account{}, err
		}
	} else {
		if err := r.run(ctx, runOptions{
			stdout: b,
			stderr: b,
			stdin:  os.Stdin,
		}, r.chainCmd.AddKeyCommand(name, coinType)); err != nil {
			return Account{}, err
		}

		data, err := b.JSONEnsuredBytes()
		if err != nil {
			return Account{}, err
		}
		if err := json.Unmarshal(data, &account); err != nil {
			return Account{}, err
		}
	}

	// 獲取賬戶地址。
	retrieved, err := r.ShowAccount(ctx, name)
	if err != nil {
		return Account{}, err
	}
	account.Address = retrieved.Address

	return account, nil
}

// ImportAccount 從密鑰文件中導入帳戶
func (r Runner) ImportAccount(ctx context.Context, name, keyFile, passphrase string) (Account, error) {
	if err := r.CheckAccountExist(ctx, name); err != nil {
		return Account{}, err
	}

// 將密碼寫為輸入
// TODO: 管理密鑰環後端而不是測試
	input := &bytes.Buffer{}
	fmt.Fprintln(input, passphrase)

	if err := r.run(
		ctx,
		runOptions{},
		r.chainCmd.ImportKeyCommand(name, keyFile),
		step.Write(input.Bytes()),
	); err != nil {
		return Account{}, err
	}

	return r.ShowAccount(ctx, name)
}

// 如果帳戶已存在於鏈密鑰環中，則 CheckAccountExist 返回錯誤
func (r Runner) CheckAccountExist(ctx context.Context, name string) error {
	b := newBuffer()

	// 獲取並解碼鏈上的所有賬戶
	var accounts []Account
	if err := r.run(ctx, runOptions{stdout: b}, r.chainCmd.ListKeysCommand()); err != nil {
		return err
	}

	data, err := b.JSONEnsuredBytes()
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &accounts); err != nil {
		return err
	}

	// 搜索帳戶名稱
	for _, account := range accounts {
		if account.Name == name {
			return ErrAccountAlreadyExists
		}
	}
	return nil
}

//ShowAccount 顯示帳戶的詳細信息。
func (r Runner) ShowAccount(ctx context.Context, name string) (Account, error) {
	b := &bytes.Buffer{}

	opt := []step.Option{
		r.chainCmd.ShowKeyAddressCommand(name),
	}

	if r.chainCmd.KeyringPassword() != "" {
		input := &bytes.Buffer{}
		fmt.Fprintln(input, r.chainCmd.KeyringPassword())
		opt = append(opt, step.Write(input.Bytes()))
	}

	if err := r.run(ctx, runOptions{stdout: b}, opt...); err != nil {
		if strings.Contains(err.Error(), "找不到項目") ||
			strings.Contains(err.Error(), "不是有效的名稱或地址") {
			return Account{}, ErrAccountDoesNotExist
		}
		return Account{}, err
	}

	return Account{
		Name:    name,
		Address: strings.TrimSpace(b.String()),
	}, nil
}

// AddGenesisAccount 通過其地址將帳戶添加到創世紀。
func (r Runner) AddGenesisAccount(ctx context.Context, address, coins string) error {
	return r.run(ctx, runOptions{}, r.chainCmd.AddGenesisAccountCommand(address, coins))
}

// AddVestingAccount 通過其地址將歸屬賬戶添加到創世紀。
func (r Runner) AddVestingAccount(
	ctx context.Context,
	address,
	originalCoins,
	vestingCoins string,
	vestingEndTime int64,
) error {
	return r.run(ctx, runOptions{}, r.chainCmd.AddVestingAccountCommand(address, originalCoins, vestingCoins, vestingEndTime))
}
