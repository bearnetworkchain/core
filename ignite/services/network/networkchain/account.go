package networkchain

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	chaincmdrunner "github.com/ignite-hq/cli/ignite/pkg/chaincmd/runner"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosutil"
	"github.com/ignite-hq/cli/ignite/pkg/randstr"
	"github.com/ignite-hq/cli/ignite/services/chain"
)

const (
	passphraseLength = 32
	sampleAccount    = "alice"
)

// InitAccount 為區塊鏈初始化一個賬戶，並在 config/gentx/gentx.json 中發出一個 gentx
func (c Chain) InitAccount(ctx context.Context, v chain.Validator, accountName string) (string, error) {
	if !c.isInitialized {
		return "", errors.New("必須初始化區塊鏈以初始化帳戶")
	}

	chainCmd, err := c.chain.Commands(ctx)
	if err != nil {
		return "", err
	}

	// 創建鏈帳戶。
	address, err := c.ImportAccount(ctx, accountName)
	if err != nil {
		return "", err
	}

	// 將帳戶添加到創世紀
	err = chainCmd.AddGenesisAccount(ctx, address, v.StakingAmount)
	if err != nil {
		return "", err
	}

	// 創建 gentx。
	issuedGentxPath, err := c.chain.IssueGentx(ctx, v)
	if err != nil {
		return "", err
	}

	// 將發出的 gentx 重命名為 gentx.json
	gentxPath := filepath.Join(filepath.Dir(issuedGentxPath), cosmosutil.GentxFilename)
	return gentxPath, os.Rename(issuedGentxPath, gentxPath)
}

// ImportAccount 將一個賬戶從 Starport 導入到鏈中。
// 我們首先將賬戶導出到臨時密鑰文件中，然後使用鏈 CLI 導入。
func (c *Chain) ImportAccount(ctx context.Context, name string) (string, error) {
// 鏈 CLI 的密鑰導入命令要求密鑰文件使用至少 8 個字符的密碼進行加密
// 我們生成一個隨機密碼來導入賬戶
	passphrase := randstr.Runes(passphraseLength)

	// 將密鑰導出到臨時文件中。
	armored, err := c.ar.Export(name, passphrase)
	if err != nil {
		return "", err
	}

	keyFile, err := os.CreateTemp("", "")
	if err != nil {
		return "", err
	}
	defer os.Remove(keyFile.Name())

	if _, err := keyFile.Write([]byte(armored)); err != nil {
		return "", err
	}

	// 將密鑰文件導入鏈中。
	chainCmd, err := c.chain.Commands(ctx)
	if err != nil {
		return "", err
	}

	acc, err := chainCmd.ImportAccount(ctx, name, keyFile.Name(), passphrase)
	return acc.Address, err
}

// detectPrefix 檢測鏈的賬戶地址前綴
// 該方法創建一個示例帳戶並從中解析地址前綴
func (c Chain) detectPrefix(ctx context.Context) (string, error) {
	chainCmd, err := c.chain.Commands(ctx)
	if err != nil {
		return "", err
	}

	var acc chaincmdrunner.Account
	acc, err = chainCmd.ShowAccount(ctx, sampleAccount)
	if errors.Is(err, chaincmdrunner.ErrAccountDoesNotExist) {
		// 示例賬戶不存在，我們創建它
		acc, err = chainCmd.AddAccount(ctx, sampleAccount, "", "")
	}
	if err != nil {
		return "", err
	}

	return cosmosutil.GetAddressPrefix(acc.Address)
}
