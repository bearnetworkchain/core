package tendermintrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	endpointNetInfo = "/net_info"
	endpointGenesis = "/genesis"
	endpointStatus  = "/status"
)

// Client是一個 Tendermint RPC 客戶端。
type Client struct {
	addr string
}

// New 創建一個新的 Tendermint RPC 客戶端。
func New(addr string) Client {
	return Client{addr: addr}
}

// NetInfo 代表網絡信息。
type NetInfo struct {
	ConnectedPeers int
}

func (c Client) url(endpoint string) string {
	return fmt.Sprintf("%s%s", c.addr, endpoint)
}

// GetNetInfo 檢索網絡信息。
func (c Client) GetNetInfo(ctx context.Context) (NetInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(endpointNetInfo), nil)
	if err != nil {
		return NetInfo{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return NetInfo{}, err
	}
	defer resp.Body.Close()

	var res struct {
		Result struct {
			Peers string `json:"n_peers"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return NetInfo{}, err
	}

	peers, err := strconv.ParseUint(res.Result.Peers, 10, 64)
	if err != nil {
		return NetInfo{}, err
	}

	return NetInfo{
		ConnectedPeers: int(peers),
	}, nil
}

// Genesis 代表創世紀。
type Genesis struct {
	ChainID string `json:"chain_id"`
}

// GetGenesis 檢索創世紀。
func (c Client) GetGenesis(ctx context.Context) (Genesis, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(endpointGenesis), nil)
	if err != nil {
		return Genesis{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Genesis{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Genesis{}, fmt.Errorf("%d", resp.StatusCode)
	}

	var out struct {
		Result struct {
			Genesis Genesis `json:"genesis"`
		} `json:"Result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Genesis{}, err
	}

	return out.Result.Genesis, nil
}

// NodeInfo 保存節點信息。
type NodeInfo struct {
	Network string
}

// Status 檢索節點狀態。
func (c Client) Status(ctx context.Context) (NodeInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(endpointStatus), nil)
	if err != nil {
		return NodeInfo{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return NodeInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return NodeInfo{}, fmt.Errorf("%d", resp.StatusCode)
	}

	var (
		info NodeInfo
		b    = &bytes.Buffer{}
		r    = io.TeeReader(resp.Body, b)
	)

	var out struct {
		Result struct {
			NodeInfo NodeInfo `json:"node_info"`
		} `json:"result"`
	}

	if err := json.NewDecoder(r).Decode(&out); err != nil {
		return NodeInfo{}, err
	}

	info = out.Result.NodeInfo

	// 一些熊網鏈版本有不同的響應負載。
	if info.Network == "" {
		var out struct {
			Result struct {
				NodeInfo NodeInfo `json:"NodeInfo"`
			} `json:"result"`
		}

		if err := json.NewDecoder(b).Decode(&out); err != nil {
			return NodeInfo{}, err
		}

		info = out.Result.NodeInfo
	}

	return info, nil
}
