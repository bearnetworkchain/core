package network

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	campaigntypes "github.com/tendermint/spn/x/campaign/types"
	launchtypes "github.com/tendermint/spn/x/launch/types"
	profiletypes "github.com/tendermint/spn/x/profile/types"

	"github.com/bearnetworkchain/core/ignite/pkg/cosmoserror"
	"github.com/bearnetworkchain/core/ignite/pkg/cosmosutil"
	"github.com/bearnetworkchain/core/ignite/pkg/events"
	"github.com/bearnetworkchain/core/ignite/services/network/networktypes"
)

// publishOptions 保存有關如何創建鏈的信息。
type publishOptions struct {
	genesisURL       string
	chainID          string
	campaignID       uint64
	noCheck          bool
	metadata         string
	totalSupply      sdk.Coins
	sharePercentages SharePercents
	mainnet          bool
}

// PublishOption 配置鏈創建。
type PublishOption func(*publishOptions)

// WithCampaign添加活動ID。
func WithCampaign(id uint64) PublishOption {
	return func(o *publishOptions) {
		o.campaignID = id
	}
}

// WithChainID 使用自定義鏈 ID。
func WithChainID(chainID string) PublishOption {
	return func(o *publishOptions) {
		o.chainID = chainID
	}
}

// WithNoCheck 禁用檢查鏈的完整性。
func WithNoCheck() PublishOption {
	return func(o *publishOptions) {
		o.noCheck = true
	}
}

// WithCustomGenesis 允許在發布期間使用自定義起源。
func WithCustomGenesis(url string) PublishOption {
	return func(o *publishOptions) {
		o.genesisURL = url
	}
}

// WithMetadata provides 更新活動的元數據提案。
func WithMetadata(metadata string) PublishOption {
	return func(c *publishOptions) {
		c.metadata = metadata
	}
}

// WithTotalSupply 提供總供應建議以更新活動。
func WithTotalSupply(totalSupply sdk.Coins) PublishOption {
	return func(c *publishOptions) {
		c.totalSupply = totalSupply
	}
}

// WithPercentageShares為股票啟用鑄造代金券。
func WithPercentageShares(sharePercentages []SharePercent) PublishOption {
	return func(c *publishOptions) {
		c.sharePercentages = sharePercentages
	}
}

// Mainnet 將已發布的鏈初始化到主網
func Mainnet() PublishOption {
	return func(o *publishOptions) {
		o.mainnet = true
	}
}

// Publish向 SPN 提交創世紀以宣布新網絡。
func (n Network) Publish(ctx context.Context, c Chain, options ...PublishOption) (launchID, campaignID uint64, err error) {
	o := publishOptions{}
	for _, apply := range options {
		apply(&o)
	}

	var (
		genesisHash string
		genesisFile []byte
		genesis     cosmosutil.ChainGenesis
	)

	// 如果初始創世是一個創世 URL 並且沒有執行檢查，我們只需獲取它並獲取它的哈希值。
	if o.genesisURL != "" {
		genesisFile, genesisHash, err = cosmosutil.GenesisAndHashFromURL(ctx, o.genesisURL)
		if err != nil {
			return 0, 0, err
		}
		genesis, err = cosmosutil.ParseChainGenesis(genesisFile)
		if err != nil {
			return 0, 0, err
		}
	}

	chainID := genesis.ChainID
	// 始終以最高優先級使用鏈 ID 標誌。
	if o.chainID != "" {
		chainID = o.chainID
	}
	// 如果鏈 id 為空，則使用默認值。
	if chainID == "" {
		chainID, err = c.ChainID()
		if err != nil {
			return 0, 0, err
		}
	}

	coordinatorAddress := n.account.Address(networktypes.SPN)
	campaignID = o.campaignID

	n.ev.Send(events.New(events.StatusOngoing, "發佈網絡"))

	_, err = n.profileQuery.
		CoordinatorByAddress(ctx, &profiletypes.QueryGetCoordinatorByAddressRequest{
			Address: coordinatorAddress,
		})
	if cosmoserror.Unwrap(err) == cosmoserror.ErrNotFound {
		msgCreateCoordinator := profiletypes.NewMsgCreateCoordinator(
			coordinatorAddress,
			"",
			"",
			"",
		)
		if _, err := n.cosmos.BroadcastTx(n.account.Name, msgCreateCoordinator); err != nil {
			return 0, 0, err
		}
	} else if err != nil {
		return 0, 0, err
	}

	if campaignID != 0 {
		_, err = n.campaignQuery.
			Campaign(ctx, &campaigntypes.QueryGetCampaignRequest{
				CampaignID: o.campaignID,
			})
		if err != nil {
			return 0, 0, err
		}
	} else {
		campaignID, err = n.CreateCampaign(c.Name(), o.metadata, o.totalSupply)
		if err != nil {
			return 0, 0, err
		}
	}

	// mint vouchers
	if !o.sharePercentages.Empty() {
		totalSharesResp, err := n.campaignQuery.TotalShares(ctx, &campaigntypes.QueryTotalSharesRequest{})
		if err != nil {
			return 0, 0, err
		}

		var coins []sdk.Coin
		for _, percentage := range o.sharePercentages {
			coin, err := percentage.Share(totalSharesResp.TotalShares)
			if err != nil {
				return 0, 0, err
			}
			coins = append(coins, coin)
		}
		// TODO 考慮遷移到 UpdateCampaign，但不確定，可能不相關。
		// 最好在一個 tx 中發送多條消息。
		// 考慮重構方法以實現更好的 API 和效率。
		msgMintVouchers := campaigntypes.NewMsgMintVouchers(
			n.account.Address(networktypes.SPN),
			campaignID,
			campaigntypes.NewSharesFromCoins(sdk.NewCoins(coins...)),
		)
		_, err = n.cosmos.BroadcastTx(n.account.Name, msgMintVouchers)
		if err != nil {
			return 0, 0, err
		}
	}

	// 根據主網標誌初始化主網或測試網
	if o.mainnet {
		launchID, err = n.InitializeMainnet(campaignID, c.SourceURL(), c.SourceHash(), chainID)
		if err != nil {
			return 0, 0, err
		}
	} else {
		msgCreateChain := launchtypes.NewMsgCreateChain(
			n.account.Address(networktypes.SPN),
			chainID,
			c.SourceURL(),
			c.SourceHash(),
			o.genesisURL,
			genesisHash,
			true,
			campaignID,
			nil,
		)
		res, err := n.cosmos.BroadcastTx(n.account.Name, msgCreateChain)
		if err != nil {
			return 0, 0, err
		}
		var createChainRes launchtypes.MsgCreateChainResponse
		if err := res.Decode(&createChainRes); err != nil {
			return 0, 0, err
		}
		launchID = createChainRes.LaunchID
	}
	if err := c.CacheBinary(launchID); err != nil {
		return 0, 0, err
	}

	return launchID, campaignID, nil
}

func (n Network) SendAccountRequestForCoordinator(launchID uint64, amount sdk.Coins) error {
	return n.sendAccountRequest(launchID, n.account.Address(networktypes.SPN), amount)
}

// SendAccountRequest 創建一個添加 AddAccount 請求消息。
func (n Network) sendAccountRequest(
	launchID uint64,
	address string,
	amount sdk.Coins,
) error {
	msg := launchtypes.NewMsgRequestAddAccount(
		n.account.Address(networktypes.SPN),
		launchID,
		address,
		amount,
	)

	n.ev.Send(events.New(events.StatusOngoing, "廣播賬戶交易"))
	res, err := n.cosmos.BroadcastTx(n.account.Name, msg)
	if err != nil {
		return err
	}

	var requestRes launchtypes.MsgRequestAddAccountResponse
	if err := res.Decode(&requestRes); err != nil {
		return err
	}

	if requestRes.AutoApproved {
		n.ev.Send(events.New(events.StatusDone, "協調者添加到網絡的帳戶!"))
	} else {
		n.ev.Send(events.New(events.StatusDone,
			fmt.Sprintf("要求 %d 添加帳戶到網絡已提交!",
				requestRes.RequestID),
		))
	}
	return nil
}
