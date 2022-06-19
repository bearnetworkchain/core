package network

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	campaigntypes "github.com/tendermint/spn/x/campaign/types"

	"github.com/ignite-hq/cli/ignite/pkg/events"
	"github.com/ignite-hq/cli/ignite/services/network/networktypes"
)

type (
	// 道具更新活動提案
	Prop func(*updateProp)

	// updateProp 表示更新活動提案
	updateProp struct {
		name        string
		metadata    []byte
		totalSupply sdk.Coins
	}
)

// WithCampaignName提供名稱建議以更新活動。
func WithCampaignName(name string) Prop {
	return func(c *updateProp) {
		c.name = name
	}
}

// WithCampaignMetadata提供元數據建議以更新活動。
func WithCampaignMetadata(metadata string) Prop {
	return func(c *updateProp) {
		c.metadata = []byte(metadata)
	}
}

// WithCampaignTotalSupply提供總供應建議以更新活動。
func WithCampaignTotalSupply(totalSupply sdk.Coins) Prop {
	return func(c *updateProp) {
		c.totalSupply = totalSupply
	}
}

// 活動從網絡中獲取活動ID
func (n Network) Campaign(ctx context.Context, campaignID uint64) (networktypes.Campaign, error) {
	n.ev.Send(events.New(events.StatusOngoing, "獲取活動信息"))
	res, err := n.campaignQuery.Campaign(ctx, &campaigntypes.QueryGetCampaignRequest{
		CampaignID: campaignID,
	})
	if err != nil {
		return networktypes.Campaign{}, err
	}
	return networktypes.ToCampaign(res.Campaign), nil
}

// Campaigns從網絡中獲取活動
func (n Network) Campaigns(ctx context.Context) ([]networktypes.Campaign, error) {
	var campaigns []networktypes.Campaign

	n.ev.Send(events.New(events.StatusOngoing, "獲取活動信息"))
	res, err := n.campaignQuery.
		CampaignAll(ctx, &campaigntypes.QueryAllCampaignRequest{})
	if err != nil {
		return campaigns, err
	}

	// 解析獲取的活動
	for _, campaign := range res.Campaign {
		campaigns = append(campaigns, networktypes.ToCampaign(campaign))
	}

	return campaigns, nil
}

// CreateCampaign在網絡中創建活動
func (n Network) CreateCampaign(name, metadata string, totalSupply sdk.Coins) (uint64, error) {
	n.ev.Send(events.New(events.StatusOngoing, fmt.Sprintf("創建活動 %s", name)))

	msgCreateCampaign := campaigntypes.NewMsgCreateCampaign(
		n.account.Address(networktypes.SPN),
		name,
		totalSupply,
		[]byte(metadata),
	)
	res, err := n.cosmos.BroadcastTx(n.account.Name, msgCreateCampaign)
	if err != nil {
		return 0, err
	}

	var createCampaignRes campaigntypes.MsgCreateCampaignResponse
	if err := res.Decode(&createCampaignRes); err != nil {
		return 0, err
	}

	return createCampaignRes.CampaignID, nil
}

// InitializeMainnet初始化活動的主網。
func (n Network) InitializeMainnet(
	campaignID uint64,
	sourceURL,
	sourceHash string,
	mainnetChainID string,
) (uint64, error) {
	n.ev.Send(events.New(events.StatusOngoing, "初始化主網活動"))
	msg := campaigntypes.NewMsgInitializeMainnet(
		n.account.Address(networktypes.SPN),
		campaignID,
		sourceURL,
		sourceHash,
		mainnetChainID,
	)

	res, err := n.cosmos.BroadcastTx(n.account.Name, msg)
	if err != nil {
		return 0, err
	}

	var initMainnetRes campaigntypes.MsgInitializeMainnetResponse
	if err := res.Decode(&initMainnetRes); err != nil {
		return 0, err
	}

	n.ev.Send(events.New(events.StatusDone, fmt.Sprintf("活動 %d 在主網上初始化", campaignID)))

	return initMainnetRes.MainnetID, nil
}

// UpdateCampaign更新活動名稱或元數據
func (n Network) UpdateCampaign(
	id uint64,
	props ...Prop,
) error {
	//應用用戶提供的選項
	p := updateProp{}
	for _, apply := range props {
		apply(&p)
	}

	n.ev.Send(events.New(events.StatusOngoing, fmt.Sprintf("更新活動 %d", id)))
	account := n.account.Address(networktypes.SPN)
	msgs := make([]sdk.Msg, 0)
	if p.name != "" || len(p.metadata) > 0 {
		msgs = append(msgs, campaigntypes.NewMsgEditCampaign(
			account,
			id,
			p.name,
			p.metadata,
		))
	}
	if !p.totalSupply.Empty() {
		msgs = append(msgs, campaigntypes.NewMsgUpdateTotalSupply(
			account,
			id,
			p.totalSupply,
		))
	}

	if _, err := n.cosmos.BroadcastTx(n.account.Name, msgs...); err != nil {
		return err
	}
	n.ev.Send(events.New(events.StatusDone, fmt.Sprintf(
		"活動 %d 更新", id,
	)))
	return nil
}
