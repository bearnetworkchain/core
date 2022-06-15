package cosmosfaucet

import (
	"context"
	"encoding/json"
	"net/http"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ignite-hq/cli/ignite/pkg/xhttp"
)

type TransferRequest struct {
	//AccountAddress 請求硬幣。
	AccountAddress string `json:"address"`

	// 請求的硬幣。
	// 未提供此選項時使用的默認選項。
	Coins []string `json:"coins"`
}

func NewTransferRequest(accountAddress string, coins []string) TransferRequest {
	return TransferRequest{
		AccountAddress: accountAddress,
		Coins:          coins,
	}
}

type TransferResponse struct {
	Error string `json:"error,omitempty"`
}

func (f Faucet) faucetHandler(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest

	// 將請求解碼為 req。
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseError(w, http.StatusBadRequest, err)
		return
	}

	// 確定要轉移的硬幣。
	coins, err := f.coinsFromRequest(req)
	if err != nil {
		responseError(w, http.StatusBadRequest, err)
		return
	}

	// 嘗試執行傳輸
	if err := f.Transfer(r.Context(), req.AccountAddress, coins); err != nil {
		if err == context.Canceled {
			return
		}
		responseError(w, http.StatusInternalServerError, err)
	} else {
		responseSuccess(w)
	}
}

// FaucetInfoResponse是水龍頭信息負載。
type FaucetInfoResponse struct {
	// IsAFaucet 表示這是一個水龍頭端點。
	// 對自動發現有用。
	IsAFaucet bool `json:"is_a_faucet"`

	// ChainID 是 chain id 水龍頭正在運行的鏈條。
	ChainID string `json:"chain_id"`
}

func (f Faucet) faucetInfoHandler(w http.ResponseWriter, r *http.Request) {
	xhttp.ResponseJSON(w, http.StatusOK, FaucetInfoResponse{
		IsAFaucet: true,
		ChainID:   f.chainID,
	})
}

// coinsFromRequest 從轉移請求中確定要轉移的代幣。
func (f Faucet) coinsFromRequest(req TransferRequest) (sdk.Coins, error) {
	if len(req.Coins) == 0 {
		return f.coins, nil
	}

	var coins []sdk.Coin
	for _, c := range req.Coins {
		coin, err := sdk.ParseCoinNormalized(c)
		if err != nil {
			return nil, err
		}
		coins = append(coins, coin)
	}

	return coins, nil
}

func responseSuccess(w http.ResponseWriter) {
	xhttp.ResponseJSON(w, http.StatusOK, TransferResponse{})
}

func responseError(w http.ResponseWriter, code int, err error) {
	xhttp.ResponseJSON(w, code, TransferResponse{
		Error: err.Error(),
	})
}
