package networktypes

import launchtypes "github.com/tendermint/spn/x/launch/types"

type (
	NetworkType string

	// ChainLaunch 代表在 SPN 上啟動一個鏈
	ChainLaunch struct {
		ID                     uint64      `json:"ID"`
		ConsumerRevisionHeight int64       `json:"ConsumerRevisionHeight"`
		ChainID                string      `json:"ChainID"`
		SourceURL              string      `json:"SourceURL"`
		SourceHash             string      `json:"SourceHash"`
		GenesisURL             string      `json:"GenesisURL"`
		GenesisHash            string      `json:"GenesisHash"`
		LaunchTime             int64       `json:"LaunchTime"`
		CampaignID             uint64      `json:"CampaignID"`
		LaunchTriggered        bool        `json:"LaunchTriggered"`
		Network                NetworkType `json:"Network"`
		Reward                 string      `json:"Reward,omitempty"`
	}
)

const (
	NetworkTypeMainnet NetworkType = "mainnet"
	NetworkTypeTestnet NetworkType = "testnet"
)

func (n NetworkType) String() string {
	return string(n)
}

// ToChainLaunch 從 SPN 轉換鏈啟動數據並返回一個 ChainLaunch 對象
func ToChainLaunch(chain launchtypes.Chain) ChainLaunch {
	var launchTime int64
	if chain.LaunchTriggered {
		launchTime = chain.LaunchTimestamp
	}

	network := NetworkTypeTestnet
	if chain.IsMainnet {
		network = NetworkTypeMainnet
	}

	launch := ChainLaunch{
		ID:                     chain.LaunchID,
		ConsumerRevisionHeight: chain.ConsumerRevisionHeight,
		ChainID:                chain.GenesisChainID,
		SourceURL:              chain.SourceURL,
		SourceHash:             chain.SourceHash,
		LaunchTime:             launchTime,
		CampaignID:             chain.CampaignID,
		LaunchTriggered:        chain.LaunchTriggered,
		Network:                network,
	}

	// 檢查是否提供了自定義創世紀 URL。
	if customGenesisURL := chain.InitialGenesis.GetGenesisURL(); customGenesisURL != nil {
		launch.GenesisURL = customGenesisURL.Url
		launch.GenesisHash = customGenesisURL.Hash
	}

	return launch
}
