package ignitecmd

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/rdegges/go-ipify"
	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/pkg/cliui/cliquiz"
	"github.com/ignite-hq/cli/ignite/pkg/cliui/icons"
	"github.com/ignite-hq/cli/ignite/pkg/gitpod"
	"github.com/ignite-hq/cli/ignite/pkg/xchisel"
	"github.com/ignite-hq/cli/ignite/services/network"
	"github.com/ignite-hq/cli/ignite/services/network/networkchain"
)

const (
	flagGentx  = "gentx"
	flagAmount = "amount"
)

// NewNetworkChainJoin 創建一個新的鏈加入命令來加入
// 作為網絡驗證者。
func NewNetworkChainJoin() *cobra.Command {
	c := &cobra.Command{
		Use:   "join [launch-id]",
		Short: "請求作為驗證者加入熊網鏈",
		Args:  cobra.ExactArgs(1),
		RunE:  networkChainJoinHandler,
	}

	c.Flags().String(flagGentx, "", "gentx json文件的路徑")
	c.Flags().String(flagAmount, "", "帳戶請求的熊網幣數量")
	c.Flags().AddFlagSet(flagNetworkFrom())
	c.Flags().AddFlagSet(flagSetHome())
	c.Flags().AddFlagSet(flagSetKeyringBackend())
	c.Flags().AddFlagSet(flagSetYes())

	return c
}

func networkChainJoinHandler(cmd *cobra.Command, args []string) error {
	session := cliui.New()
	defer session.Cleanup()

	var (
		gentxPath, _ = cmd.Flags().GetString(flagGentx)
		amount, _    = cmd.Flags().GetString(flagAmount)
	)

	nb, err := newNetworkBuilder(cmd, CollectEvents(session.EventBus()))
	if err != nil {
		return err
	}

	// parse launch ID.
	launchID, err := network.ParseID(args[0])
	if err != nil {
		return err
	}

	joinOptions := []network.JoinOption{
		network.WithCustomGentxPath(gentxPath),
	}

	// 如果沒有自定義gentx，我們需要檢測公共地址。
	if gentxPath == "" {
		// 獲取驗證者的對等公共地址。
		publicAddr, err := askPublicAddress(cmd.Context(), session)
		if err != nil {
			return err
		}

		joinOptions = append(joinOptions, network.WithPublicAddress(publicAddr))
	}

	n, err := nb.Network()
	if err != nil {
		return err
	}

	chainLaunch, err := n.ChainLaunch(cmd.Context(), launchID)
	if err != nil {
		return err
	}

	c, err := nb.Chain(networkchain.SourceLaunch(chainLaunch))
	if err != nil {
		return err
	}

	if amount != "" {
		// 解析金額。
		amountCoins, err := sdk.ParseCoinsNormalized(amount)
		if err != nil {
			return errors.Wrap(err, "錯誤解析量")
		}
		joinOptions = append(joinOptions, network.WithAccountRequest(amountCoins))
	} else {
		if !getYes(cmd) {
			question := fmt.Sprintf(
				"你沒有設置 --%s flag 因此不會提交帳戶請求。 請您確認",
				flagAmount,
			)
			if err := session.AskConfirm(question); err != nil {
				return session.PrintSaidNo()
			}
		}

		session.Printf("%s %s\n", icons.Info, "不會提交帳戶請求")
	}

	// 創建消息以添加驗證器。
	return n.Join(cmd.Context(), c, launchID, joinOptions...)
}

// askPublicAddress 準備問題以交互方式詢問 publicAddress
// 當未提供對等點且未通過鑿子代理運行時。
func askPublicAddress(ctx context.Context, session cliui.Session) (publicAddress string, err error) {
	options := []cliquiz.Option{
		cliquiz.Required(),
	}
	if gitpod.IsOnGitpod() {
		publicAddress, err = gitpod.URLForPort(ctx, xchisel.DefaultServerPort)
		if err != nil {
			return "", errors.Wrap(err, "無法讀取節點的公共 Gitpod 地址")
		}
		return publicAddress, nil
	}

	// 即使 GetIp 失敗，我們也不會處理錯誤，因為我們不想中斷連接過程。
	// 萬一GetIp失敗，用戶應該手動輸入他的地址
	ip, err := ipify.GetIp()
	if err == nil {
		options = append(options, cliquiz.DefaultAnswer(fmt.Sprintf("%s:26656", ip)))
	}

	questions := []cliquiz.Question{cliquiz.NewQuestion(
		"同行的地址",
		&publicAddress,
		options...,
	)}
	return publicAddress, session.Ask(questions...)
}
