package networkchain

import (
	"path/filepath"

	"github.com/ignite-hq/cli/ignite/pkg/confile"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosutil"
)

const (
	HTTPTunnelChisel = "chisel"
)

const SPNConfigFile = "spn.yml"

type Config struct {
	TunneledPeers []TunneledPeer `json:"tunneled_peers" yaml:"tunneled_peers"`
}

//TunneledPeer 代表無法通過常規 tcp 連接到達的對等點的 http 隧道
type TunneledPeer struct {
	// 名稱代表隧道類型，例如“隧道”
	Name string `json:"name" yaml:"name"`

	// 地址代表隧道的http地址，例如. "https://tendermint-starport-i5e75cplx02.ws-eu31.gitpod.io/"
	Address string `json:"address" yaml:"address"`

	// NodeID 隧道後面的節點的節點 ID，例如 "e6a59e37b2761f26a21c9168f78a7f2b07c120c7"
	NodeID string `json:"node_id" yaml:"node_id"`

	// LocalPort 指定必須用於本地隧道客戶端的端口
	LocalPort string `json:"local_port" yaml:"local_port"`
}

func GetSPNConfig(path string) (conf Config, err error) {
	err = confile.New(confile.DefaultYAMLEncodingCreator, path).Load(&conf)
	return
}

func SetSPNConfig(config Config, path string) error {
	return confile.New(confile.DefaultYAMLEncodingCreator, path).Save(config)
}

func (c *Chain) SPNConfigPath() (string, error) {
	home, err := c.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, cosmosutil.ChainConfigDir, SPNConfigFile), nil
}
