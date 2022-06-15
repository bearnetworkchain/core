package cosmosfaucet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

// ErrTransferRequest 是傳輸請求失敗時發生的錯誤
type ErrTransferRequest struct {
	StatusCode int
}

// 錯誤實施錯誤
func (err ErrTransferRequest) Error() string {
	return http.StatusText(err.StatusCode)
}

// HTTPClient 是一個水龍頭客戶端。
type HTTPClient struct {
	addr string
}

// NewClient 返回一個新的水龍頭客戶端。
func NewClient(addr string) HTTPClient {
	return HTTPClient{addr}
}

// 使用 req 從水龍頭轉移請求令牌。
func (c HTTPClient) Transfer(ctx context.Context, req TransferRequest) (TransferResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return TransferResponse{}, err
	}

	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.addr, bytes.NewReader(data))
	if err != nil {
		return TransferResponse{}, err
	}

	hres, err := http.DefaultClient.Do(hreq)
	if err != nil {
		return TransferResponse{}, err
	}
	defer hres.Body.Close()

	if hres.StatusCode != http.StatusOK {
		return TransferResponse{}, ErrTransferRequest{hres.StatusCode}
	}

	var res TransferResponse
	err = json.NewDecoder(hres.Body).Decode(&res)
	return res, err
}

// FaucetInfo 為客戶端獲取水龍頭信息以確定這是否是一個真正的水龍頭和
// 水龍頭正在運行的鏈的鏈 ID 是什麼。
func (c HTTPClient) FaucetInfo(ctx context.Context) (FaucetInfoResponse, error) {
	hreq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.addr+"/info", nil)
	if err != nil {
		return FaucetInfoResponse{}, err
	}

	hres, err := http.DefaultClient.Do(hreq)
	if err != nil {
		return FaucetInfoResponse{}, err
	}
	defer hres.Body.Close()

	if hres.StatusCode != http.StatusOK {
		return FaucetInfoResponse{}, errors.New(http.StatusText(hres.StatusCode))
	}

	var res FaucetInfoResponse
	err = json.NewDecoder(hres.Body).Decode(&res)
	return res, err
}
