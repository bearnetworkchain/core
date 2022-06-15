package networktypes

import (
	"fmt"

	launchtypes "github.com/tendermint/spn/x/launch/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/ignite-hq/cli/ignite/pkg/cosmosutil"
	"github.com/ignite-hq/cli/ignite/pkg/xtime"
)

type (
	//Request 表示 SPN 上一條鏈的啟動請求
	Request struct {
		LaunchID  uint64                     `json:"LaunchID"`
		RequestID uint64                     `json:"RequestID"`
		Creator   string                     `json:"Creator"`
		CreatedAt string                     `json:"CreatedAt"`
		Content   launchtypes.RequestContent `json:"Content"`
		Status    string                     `json:"Status"`
	}
)

// ToRequest 從 SPN 轉換一個請求數據並返回一個 Request 對象
func ToRequest(request launchtypes.Request) Request {

	return Request{
		LaunchID:  request.LaunchID,
		RequestID: request.RequestID,
		Creator:   request.Creator,
		CreatedAt: xtime.FormatUnixInt(request.CreatedAt),
		Content:   request.Content,
		Status:    launchtypes.Request_Status_name[int32(request.Status)],
	}
}

// VerifyRequest 從其內容中驗證請求的有效性（靜態檢查）
func VerifyRequest(request Request) error {
	req, ok := request.Content.Content.(*launchtypes.RequestContent_GenesisValidator)
	if ok {
		err := VerifyAddValidatorRequest(req)
		if err != nil {
			return NewWrappedErrInvalidRequest(request.RequestID, err.Error())
		}
	}

	return nil
}

// VerifyAddValidatorRequest 驗證驗證器請求參數
func VerifyAddValidatorRequest(req *launchtypes.RequestContent_GenesisValidator) error {
// 如果這是一個添加驗證器請求
	var (
		peer           = req.GenesisValidator.Peer
		valAddress     = req.GenesisValidator.Address
		consPubKey     = req.GenesisValidator.ConsPubKey
		selfDelegation = req.GenesisValidator.SelfDelegation
	)

	// Check values inside the gentx are correct
	info, _, err := cosmosutil.ParseGentx(req.GenesisValidator.GenTx)
	if err != nil {
		return fmt.Errorf("無法解析 gentx %s", err.Error())
	}

// 將從 gentx 獲取的地址前綴更改為 SPN 上使用的地址前綴
// 因為 SPN 上的所有鏈上存儲地址都使用 SPN 前綴
	spnFetchedAddress, err := cosmosutil.ChangeAddressPrefix(info.DelegatorAddress, SPN)
	if err != nil {
		return err
	}

	//檢查驗證者地址
	if valAddress != spnFetchedAddress {
		return fmt.Errorf(
			"驗證者地址 %s 與 gentx 內部的不匹配 %s",
			valAddress,
			spnFetchedAddress,
		)
	}

	// 檢查驗證者地址
	if !info.PubKey.Equals(ed25519.PubKey(consPubKey)) {
		return fmt.Errorf(
			"共識公鑰 %s 與 gentx 內部的不匹配 %s",
			ed25519.PubKey(consPubKey).String(),
			info.PubKey.String(),
		)
	}

	// 檢查自我委託
	if selfDelegation.Denom != info.SelfDelegation.Denom ||
		!selfDelegation.IsEqual(info.SelfDelegation) {
		return fmt.Errorf(
			"自我委託 %s 與 gentx 內部的不匹配 %s",
			selfDelegation.String(),
			info.SelfDelegation.String(),
		)
	}

	//檢查對等體的格式
	if !cosmosutil.VerifyPeerFormat(peer) {
		return fmt.Errorf(
			"對等地址 %s 與對等格式不匹配 <host>:<port>",
			peer.String(),
		)
	}
	return nil
}
