package ignitecmd

import (
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tendermint/spn/pkg/chainid"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/pkg/cliui/icons"
	"github.com/ignite-hq/cli/ignite/pkg/xurl"
	"github.com/ignite-hq/cli/ignite/services/network"
	"github.com/ignite-hq/cli/ignite/services/network/networkchain"
)

const (
	flagTag          = "tag"
	flagBranch       = "branch"
	flagHash         = "hash"
	flagGenesis      = "genesis"
	flagCampaign     = "campaign"
	flagShares       = "shares"
	flagNoCheck      = "no-check"
	flagChainID      = "chain-id"
	flagMainnet      = "mainnet"
	flagRewardCoins  = "reward.coins"
	flagRewardHeight = "reward.height"
)

// NewNetworkChainPublish 返回一個新命令來發布一條新鏈以啟動一個新網絡。
func NewNetworkChainPublish() *cobra.Command {
	c := &cobra.Command{
		Use:   "publish [source-url]",
		Short: "發布新鏈啟動新網絡",
		Args:  cobra.ExactArgs(1),
		RunE:  networkChainPublishHandler,
	}

	flagSetClearCache(c)
	c.Flags().String(flagBranch, "", "用於 repo 的 Git 分支")
	c.Flags().String(flagTag, "", "用於 repo 的 Git 標記")
	c.Flags().String(flagHash, "", "用於存儲庫的 Git 哈希")
	c.Flags().String(flagGenesis, "", "自定義創世紀的 URL")
	c.Flags().String(flagChainID, "", "用於此網絡的Chain ID")
	c.Flags().Uint64(flagCampaign, 0, "用於此網絡的活動 ID")
	c.Flags().Bool(flagNoCheck, false, "跳過驗證鏈的完整性")
	c.Flags().String(flagCampaignMetadata, "", "添加活動元數據")
	c.Flags().String(flagCampaignTotalSupply, "", "添加一個活動的總主網")
	c.Flags().String(flagShares, "", "為活動添加分享")
	c.Flags().Bool(flagMainnet, false, "初始化主網活動")
	c.Flags().String(flagRewardCoins, "", "獎勵金幣")
	c.Flags().Int64(flagRewardHeight, 0, "最後獎勵高度")
	c.Flags().String(flagAmount, "", "帳戶請求的熊網幣數量")
	c.Flags().AddFlagSet(flagNetworkFrom())
	c.Flags().AddFlagSet(flagSetKeyringBackend())
	c.Flags().AddFlagSet(flagSetHome())
	c.Flags().AddFlagSet(flagSetYes())

	return c
}

func networkChainPublishHandler(cmd *cobra.Command, args []string) error {
	session := cliui.New()
	defer session.Cleanup()

	var (
		tag, _                    = cmd.Flags().GetString(flagTag)
		branch, _                 = cmd.Flags().GetString(flagBranch)
		hash, _                   = cmd.Flags().GetString(flagHash)
		genesisURL, _             = cmd.Flags().GetString(flagGenesis)
		chainID, _                = cmd.Flags().GetString(flagChainID)
		campaign, _               = cmd.Flags().GetUint64(flagCampaign)
		noCheck, _                = cmd.Flags().GetBool(flagNoCheck)
		campaignMetadata, _       = cmd.Flags().GetString(flagCampaignMetadata)
		campaignTotalSupplyStr, _ = cmd.Flags().GetString(flagCampaignTotalSupply)
		sharesStr, _              = cmd.Flags().GetString(flagShares)
		isMainnet, _              = cmd.Flags().GetBool(flagMainnet)
		rewardCoinsStr, _         = cmd.Flags().GetString(flagRewardCoins)
		rewardDuration, _         = cmd.Flags().GetInt64(flagRewardHeight)
		amount, _                 = cmd.Flags().GetString(flagAmount)
	)

	// parse the amount.
	amountCoins, err := sdk.ParseCoinsNormalized(amount)
	if err != nil {
		return errors.Wrap(err, "金額解析錯誤")
	}

	source, err := xurl.MightHTTPS(args[0])
	if err != nil {
		return fmt.Errorf("無效的URL格式: %w", err)
	}

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	if campaign != 0 && campaignTotalSupplyStr != "" {
		return fmt.Errorf("%s 和 %s flags 不能一起設置", flagCampaign, flagCampaignTotalSupply)
	}
	if isMainnet {
		if campaign == 0 && campaignTotalSupplyStr == "" {
			return fmt.Errorf(
				"%s flag 需要其中之一 %s or %s flags 要設置",
				flagMainnet,
				flagCampaign,
				flagCampaignTotalSupply,
			)
		}
		if chainID == "" {
			return fmt.Errorf("%s flag 需要 %s flag", flagMainnet, flagChainID)
		}
	}

	if chainID != "" {
		chainName, _, err := chainid.ParseGenesisChainID(chainID)
		if err != nil {
			return errors.Wrapf(err, "無效的 chain id: %s", chainID)
		}
		if err := chainid.CheckChainName(chainName); err != nil {
			return errors.Wrapf(err, "無效的 chain id 名稱: %s", chainName)
		}
	}

	totalSupply, err := sdk.ParseCoinsNormalized(campaignTotalSupplyStr)
	if err != nil {
		return err
	}

	rewardCoins, err := sdk.ParseCoinsNormalized(rewardCoinsStr)
	if err != nil {
		return err
	}

	if (!rewardCoins.Empty() && rewardDuration == 0) ||
		(rewardCoins.Empty() && rewardDuration > 0) {
		return fmt.Errorf("%s 和 %s flags 必須一起提供", flagRewardCoins, flagRewardHeight)
	}

	nb, err := newNetworkBuilder(cmd, CollectEvents(session.EventBus()))
	if err != nil {
		return err
	}

	// 使用來自所選目標的源。
	var sourceOption networkchain.SourceOption

	switch {
	case tag != "":
		sourceOption = networkchain.SourceRemoteTag(source, tag)
	case branch != "":
		sourceOption = networkchain.SourceRemoteBranch(source, branch)
	case hash != "":
		sourceOption = networkchain.SourceRemoteHash(source, hash)
	default:
		sourceOption = networkchain.SourceRemote(source)
	}

	var initOptions []networkchain.Option

	// 如果給定，則使用來自 url 的自定義起源。
	if genesisURL != "" {
		initOptions = append(initOptions, networkchain.WithGenesisFromURL(genesisURL))
	}

	// 在臨時目錄中初始化。
	homeDir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(homeDir)

	initOptions = append(initOptions, networkchain.WithHome(homeDir))

	// 準備發布選項
	publishOptions := []network.PublishOption{network.WithMetadata(campaignMetadata)}

	if genesisURL != "" {
		publishOptions = append(publishOptions, network.WithCustomGenesis(genesisURL))
	}

	if campaign != 0 {
		publishOptions = append(publishOptions, network.WithCampaign(campaign))
	} else if campaignTotalSupplyStr != "" {
		totalSupply, err := sdk.ParseCoinsNormalized(campaignTotalSupplyStr)
		if err != nil {
			return err
		}
		if !totalSupply.Empty() {
			publishOptions = append(publishOptions, network.WithTotalSupply(totalSupply))
		}
	}

	// 如果給定，則使用自定義Chain ID。
	if chainID != "" {
		publishOptions = append(publishOptions, network.WithChainID(chainID))
	}

	if isMainnet {
		publishOptions = append(publishOptions, network.Mainnet())
	}

	if !totalSupply.Empty() {
		publishOptions = append(publishOptions, network.WithTotalSupply(totalSupply))
	}

	if sharesStr != "" {
		sharePercentages, err := network.ParseSharePercents(sharesStr)
		if err != nil {
			return err
		}

		publishOptions = append(publishOptions, network.WithPercentageShares(sharePercentages))
	}

	// 初始化鏈.
	c, err := nb.Chain(sourceOption, initOptions...)
	if err != nil {
		return err
	}

	if noCheck {
		publishOptions = append(publishOptions, network.WithNoCheck())
	} else if err := c.Init(cmd.Context(), cacheStorage); err != nil { // 初始化鏈以進行檢查。
		return err
	}

	session.StartSpinner("發佈中...")

	n, err := nb.Network()
	if err != nil {
		return err
	}

	launchID, campaignID, err := n.Publish(cmd.Context(), c, publishOptions...)
	if err != nil {
		return err
	}

	if !rewardCoins.IsZero() && rewardDuration > 0 {
		if err := n.SetReward(launchID, rewardDuration, rewardCoins); err != nil {
			return err
		}
	}

	if !amountCoins.IsZero() {
		if err := n.SendAccountRequestForCoordinator(launchID, amountCoins); err != nil {
			return err
		}
	}

	session.StopSpinner()
	session.Printf("%s 網絡發佈 \n", icons.OK)
	if isMainnet {
		session.Printf("%s 主網 ID: %d \n", icons.Bullet, launchID)
	} else {
		session.Printf("%s 啟動 ID: %d \n", icons.Bullet, launchID)
	}
	session.Printf("%s 活動 ID: %d \n", icons.Bullet, campaignID)

	return nil
}
