package ignitecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/pkg/cliui/cliquiz"
	"github.com/ignite-hq/cli/ignite/pkg/cliui/icons"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosaccount"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosutil"
	"github.com/ignite-hq/cli/ignite/services/chain"
	"github.com/ignite-hq/cli/ignite/services/network"
	"github.com/ignite-hq/cli/ignite/services/network/networkchain"
)

const (
	flagValidatorAccount         = "validator-account"
	flagValidatorWebsite         = "validator-website"
	flagValidatorDetails         = "validator-details"
	flagValidatorSecurityContact = "validator-security-contact"
	flagValidatorMoniker         = "validator-moniker"
	flagValidatorIdentity        = "validator-identity"
	flagValidatorSelfDelegation  = "validator-self-delegation"
	flagValidatorGasPrice        = "validator-gas-price"
)

// 新網絡鏈初始化返回一個新命令以從已發布的鏈 ID 初始化鏈
func NewNetworkChainInit() *cobra.Command {
	c := &cobra.Command{
		Use:   "init [launch-id]",
		Short: "初始化已經發佈的chain-id",
		Args:  cobra.ExactArgs(1),
		RunE:  networkChainInitHandler,
	}

	flagSetClearCache(c)
	c.Flags().String(flagValidatorAccount, cosmosaccount.DefaultAccount, "熊網鏈鏈驗證者帳戶")
	c.Flags().String(flagValidatorWebsite, "", "將網站與驗證器關聯")
	c.Flags().String(flagValidatorDetails, "", "有關驗證器的詳細信息")
	c.Flags().String(flagValidatorSecurityContact, "", "驗證者安全聯繫人電子郵件")
	c.Flags().String(flagValidatorMoniker, "", "自定義驗證器對象名字")
	c.Flags().String(flagValidatorIdentity, "", "驗證者身份簽名（例如端口UPort或密鑰庫Keybase)")
	c.Flags().String(flagValidatorSelfDelegation, "", "驗證者最小自我委託")
	c.Flags().String(flagValidatorGasPrice, "", "驗證者 gas 價格")
	c.Flags().AddFlagSet(flagNetworkFrom())
	c.Flags().AddFlagSet(flagSetHome())
	c.Flags().AddFlagSet(flagSetKeyringBackend())
	c.Flags().AddFlagSet(flagSetYes())
	return c
}

func networkChainInitHandler(cmd *cobra.Command, args []string) error {
	session := cliui.New()
	defer session.Cleanup()

	nb, err := newNetworkBuilder(cmd, CollectEvents(session.EventBus()))
	if err != nil {
		return err
	}

	// 解析啟動 ID
	launchID, err := network.ParseID(args[0])
	if err != nil {
		return err
	}

	// 檢查為驗證者提供的帳戶是否存在。
	validatorAccount, _ := cmd.Flags().GetString(flagValidatorAccount)
	if _, err = nb.AccountRegistry.GetByName(validatorAccount); err != nil {
		return err
	}

	// 如果一個鏈已經用這個啟動 ID 初始化，我們請求確認
	// 在刪除目錄之前。
	chainHome, exist, err := networkchain.IsChainHomeExist(launchID)
	if err != nil {
		return err
	}

	if !getYes(cmd) && exist {
		question := fmt.Sprintf(
			"網鏈已經在下面初始化: %s. 是否要覆蓋主目錄",
			chainHome,
		)
		if err := session.AskConfirm(question); err != nil {
			return session.PrintSaidNo()
		}
	}

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
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

	if err := c.Init(cmd.Context(), cacheStorage); err != nil {
		return err
	}

	genesisPath, err := c.GenesisPath()
	if err != nil {
		return err
	}

	genesis, err := cosmosutil.ParseGenesisFromPath(genesisPath)
	if err != nil {
		return err
	}

	// 詢問驗證者信息。
	v, err := askValidatorInfo(cmd, session, genesis.StakeDenom)
	if err != nil {
		return err
	}
	session.StartSpinner("生成您的Gentx")

	gentxPath, err := c.InitAccount(cmd.Context(), v, validatorAccount)
	if err != nil {
		return err
	}

	session.StopSpinner()

	return session.Printf("%s Gentx 生成: %s\n", icons.Bullet, gentxPath)
}

// askValidatorInfo 提示用戶問題以查詢驗證器信息
func askValidatorInfo(cmd *cobra.Command, session cliui.Session, stakeDenom string) (chain.Validator, error) {
	var (
		account, _         = cmd.Flags().GetString(flagValidatorAccount)
		website, _         = cmd.Flags().GetString(flagValidatorWebsite)
		details, _         = cmd.Flags().GetString(flagValidatorDetails)
		securityContact, _ = cmd.Flags().GetString(flagValidatorSecurityContact)
		moniker, _         = cmd.Flags().GetString(flagValidatorMoniker)
		identity, _        = cmd.Flags().GetString(flagValidatorIdentity)
		selfDelegation, _  = cmd.Flags().GetString(flagValidatorSelfDelegation)
		gasPrice, _        = cmd.Flags().GetString(flagValidatorGasPrice)
	)
	if gasPrice == "" {
		gasPrice = "0" + stakeDenom
	}
	v := chain.Validator{
		Name:              account,
		Website:           website,
		Details:           details,
		Moniker:           moniker,
		Identity:          identity,
		SecurityContact:   securityContact,
		MinSelfDelegation: selfDelegation,
		GasPrices:         gasPrice,
	}

	questions := append([]cliquiz.Question{},
		cliquiz.NewQuestion("質押金額",
			&v.StakingAmount,
			cliquiz.DefaultAnswer("168888bnkt"),
			cliquiz.Required(),
		),
		cliquiz.NewQuestion("佣金率",
			&v.CommissionRate,
			cliquiz.DefaultAnswer("0.10"),
			cliquiz.Required(),
		),
		cliquiz.NewQuestion("佣金最高費率",
			&v.CommissionMaxRate,
			cliquiz.DefaultAnswer("0.20"),
			cliquiz.Required(),
		),
		cliquiz.NewQuestion("佣金最大變化率",
			&v.CommissionMaxChangeRate,
			cliquiz.DefaultAnswer("0.01"),
			cliquiz.Required(),
		),
	)
	return v, session.Ask(questions...)
}
