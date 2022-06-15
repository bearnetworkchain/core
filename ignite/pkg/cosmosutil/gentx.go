package cosmosutil

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var GentxFilename = "gentx.json"

type (
	// GentxInfo 表示關於 gentx 文件的基本信息
	GentxInfo struct {
		DelegatorAddress string
		PubKey           ed25519.PubKey
		SelfDelegation   sdk.Coin
		Memo             string
	}

	// StargateGentx 代表熊網鏈gentx文件
	StargateGentx struct {
		Body struct {
			Messages []struct {
				DelegatorAddress string `json:"delegator_address"`
				ValidatorAddress string `json:"validator_address"`
				PubKey           struct {
					Type string `json:"@type"`
					Key  string `json:"key"`
				} `json:"pubkey"`
				Value struct {
					Denom  string `json:"denom"`
					Amount string `json:"amount"`
				} `json:"value"`
			} `json:"messages"`
			Memo string `json:"memo"`
		} `json:"body"`
	}
)

// GentxFromPath 從 json 文件返回 GentxInfo
func GentxFromPath(path string) (info GentxInfo, gentx []byte, err error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return info, gentx, errors.New("鏈的主文件夾尚未初始化: " + path)
	}

	gentx, err = os.ReadFile(path)
	if err != nil {
		return info, gentx, err
	}
	return ParseGentx(gentx)
}

// ParseGentx 以字節為單位返回 GentxInfo 和 gentx 文件
// TODO 反射器。無需返回文件，它已經在參數中作為 gentx 給出.
func ParseGentx(gentx []byte) (info GentxInfo, file []byte, err error) {
	// Try parsing Stargate gentx
	var stargateGentx StargateGentx
	if err := json.Unmarshal(gentx, &stargateGentx); err != nil {
		return info, gentx, err
	}
	if stargateGentx.Body.Messages == nil {
		return info, gentx, errors.New("gentx 無法解析")
	}

	// This is a stargate gentx
	if len(stargateGentx.Body.Messages) != 1 {
		return info, gentx, errors.New("添加驗證器 gentx 必須包含 1 條消息")
	}

	info.Memo = stargateGentx.Body.Memo
	info.DelegatorAddress = stargateGentx.Body.Messages[0].DelegatorAddress

	pb := stargateGentx.Body.Messages[0].PubKey.Key
	info.PubKey, err = base64.StdEncoding.DecodeString(pb)
	if err != nil {
		return info, gentx, fmt.Errorf("無效的驗證者公鑰 %s", err.Error())
	}

	amount, ok := sdk.NewIntFromString(stargateGentx.Body.Messages[0].Value.Amount)
	if !ok {
		return info, gentx, errors.New("gentx內部的自委託無效")
	}

	info.SelfDelegation = sdk.NewCoin(
		stargateGentx.Body.Messages[0].Value.Denom,
		amount,
	)

	return info, gentx, nil
}
