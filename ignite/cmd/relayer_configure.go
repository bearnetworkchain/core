package ignitecmd

import (
	"github.com/gookit/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/pkg/cliui"
	"github.com/bearnetworkchain/core/ignite/pkg/cliui/cliquiz"
	"github.com/bearnetworkchain/core/ignite/pkg/cliui/entrywriter"
	"github.com/bearnetworkchain/core/ignite/pkg/cosmosaccount"
	"github.com/bearnetworkchain/core/ignite/pkg/relayer"
	relayerconfig "github.com/bearnetworkchain/core/ignite/pkg/relayer/config"
)

const (
	flagAdvanced            = "advanced"
	flagSourceAccount       = "source-account"
	flagTargetAccount       = "target-account"
	flagSourceRPC           = "source-rpc"
	flagTargetRPC           = "target-rpc"
	flagSourceFaucet        = "source-faucet"
	flagTargetFaucet        = "target-faucet"
	flagSourcePort          = "source-port"
	flagSourceVersion       = "source-version"
	flagTargetPort          = "target-port"
	flagTargetVersion       = "target-version"
	flagSourceGasPrice      = "source-gasprice"
	flagTargetGasPrice      = "target-gasprice"
	flagSourceGasLimit      = "source-gaslimit"
	flagTargetGasLimit      = "target-gaslimit"
	flagSourceAddressPrefix = "source-prefix"
	flagTargetAddressPrefix = "target-prefix"
	flagOrdered             = "ordered"
	flagReset               = "reset"
	flagSourceClientID      = "source-client-id"
	flagTargetClientID      = "target-client-id"

	relayerSource = "source"
	relayerTarget = "target"

	defaultSourceRPCAddress = "http://localhost:26657"
	defaultTargetRPCAddress = "https://rpc.cosmos.network:443"

	defautSourceGasPrice      = "0.00025bnkt"
	defautTargetGasPrice      = "0.025ubnkt"
	defautSourceGasLimit      = 300000
	defautTargetGasLimit      = 300000
	defautSourceAddressPrefix = "bnkt"
	defautTargetAddressPrefix = "bnkt"
)

// NewRelayerConfigure 返回一個新的中繼器配置命令。
// 水龍頭地址是可選的，連接命令將嘗試猜測地址
// 未提供時。即使自動檢索硬幣失敗，連接命令也會成功完成。
func NewRelayerConfigure() *cobra.Command {
	c := &cobra.Command{
		Use:     "configure",
		Short:   "配置源鏈和目標鏈以進行中繼",
		Aliases: []string{"conf"},
		RunE:    relayerConfigureHandler,
	}

	c.Flags().BoolP(flagAdvanced, "a", false, "自定義 IBC 模塊的高級配置選項")
	c.Flags().String(flagSourceRPC, "", "源鏈的RPC地址")
	c.Flags().String(flagTargetRPC, "", "目標鏈的RPC地址")
	c.Flags().String(flagSourceFaucet, "", "源鏈的水龍頭地址")
	c.Flags().String(flagTargetFaucet, "", "目標鏈的水龍頭地址")
	c.Flags().String(flagSourcePort, "", "源鏈上的 IBC 端口 ID")
	c.Flags().String(flagSourceVersion, "", "源鏈上的模塊版本")
	c.Flags().String(flagTargetPort, "", "目標鏈上的 IBC 端口 ID")
	c.Flags().String(flagTargetVersion, "", "目標鏈上的模塊版本")
	c.Flags().String(flagSourceGasPrice, "", "用於源鏈交易的 Gas 價格")
	c.Flags().String(flagTargetGasPrice, "", "用於目標鏈上交易的 Gas 價格")
	c.Flags().Int64(flagSourceGasLimit, 0, "用於源鏈上交易的氣體限制")
	c.Flags().Int64(flagTargetGasLimit, 0, "用於目標鏈上交易的氣體限制")
	c.Flags().String(flagSourceAddressPrefix, "", "源鏈地址前綴")
	c.Flags().String(flagTargetAddressPrefix, "", "目標鏈的地址前綴")
	c.Flags().String(flagSourceAccount, "", "來源賬戶")
	c.Flags().String(flagTargetAccount, "", "目標賬戶")
	c.Flags().Bool(flagOrdered, false, "按順序設置頻道")
	c.Flags().BoolP(flagReset, "r", false, "重置中繼器配置")
	c.Flags().String(flagSourceClientID, "", "使用自定義客戶端 ID 作為源")
	c.Flags().String(flagTargetClientID, "", "為目標使用自定義客戶端 ID")
	c.Flags().AddFlagSet(flagSetKeyringBackend())

	return c
}

func relayerConfigureHandler(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		err = handleRelayerAccountErr(err)
	}()

	session := cliui.New()
	defer session.Cleanup()

	ca, err := cosmosaccount.New(
		cosmosaccount.WithKeyringBackend(getKeyringBackend(cmd)),
	)
	if err != nil {
		return err
	}

	if err := ca.EnsureDefaultAccount(); err != nil {
		return err
	}

	if err := printSection(session, "Setting up chains"); err != nil {
		return err
	}

	// basic configuration
	var (
		sourceAccount       string
		targetAccount       string
		sourceRPCAddress    string
		targetRPCAddress    string
		sourceFaucetAddress string
		targetFaucetAddress string
		sourceGasPrice      string
		targetGasPrice      string
		sourceGasLimit      int64
		targetGasLimit      int64
		sourceAddressPrefix string
		targetAddressPrefix string
	)

	// 通道的高級配置
	var (
		sourcePort    string
		sourceVersion string
		targetPort    string
		targetVersion string
	)

	// 問題
	var (
		questionSourceAccount = cliquiz.NewQuestion(
			"來源賬戶",
			&sourceAccount,
			cliquiz.DefaultAnswer(cosmosaccount.DefaultAccount),
			cliquiz.Required(),
		)
		questionTargetAccount = cliquiz.NewQuestion(
			"目標賬戶",
			&targetAccount,
			cliquiz.DefaultAnswer(cosmosaccount.DefaultAccount),
			cliquiz.Required(),
		)
		questionSourceRPCAddress = cliquiz.NewQuestion(
			"源 RPC",
			&sourceRPCAddress,
			cliquiz.DefaultAnswer(defaultSourceRPCAddress),
			cliquiz.Required(),
		)
		questionSourceFaucet = cliquiz.NewQuestion(
			"源頭水龍頭",
			&sourceFaucetAddress,
		)
		questionTargetRPCAddress = cliquiz.NewQuestion(
			"目標 RPC",
			&targetRPCAddress,
			cliquiz.DefaultAnswer(defaultTargetRPCAddress),
			cliquiz.Required(),
		)
		questionTargetFaucet = cliquiz.NewQuestion(
			"目標水龍頭",
			&targetFaucetAddress,
		)
		questionSourcePort = cliquiz.NewQuestion(
			"源端口",
			&sourcePort,
			cliquiz.DefaultAnswer(relayer.TransferPort),
			cliquiz.Required(),
		)
		questionSourceVersion = cliquiz.NewQuestion(
			"源版本",
			&sourceVersion,
			cliquiz.DefaultAnswer(relayer.TransferVersion),
			cliquiz.Required(),
		)
		questionTargetPort = cliquiz.NewQuestion(
			"目標端口",
			&targetPort,
			cliquiz.DefaultAnswer(relayer.TransferPort),
			cliquiz.Required(),
		)
		questionTargetVersion = cliquiz.NewQuestion(
			"目標版本",
			&targetVersion,
			cliquiz.DefaultAnswer(relayer.TransferVersion),
			cliquiz.Required(),
		)
		questionSourceGasPrice = cliquiz.NewQuestion(
			"源Gas價格",
			&sourceGasPrice,
			cliquiz.DefaultAnswer(defautSourceGasPrice),
			cliquiz.Required(),
		)
		questionTargetGasPrice = cliquiz.NewQuestion(
			"目標Gas價格",
			&targetGasPrice,
			cliquiz.DefaultAnswer(defautTargetGasPrice),
			cliquiz.Required(),
		)
		questionSourceGasLimit = cliquiz.NewQuestion(
			"源氣體限制 Gas Limit",
			&sourceGasLimit,
			cliquiz.DefaultAnswer(defautSourceGasLimit),
			cliquiz.Required(),
		)
		questionTargetGasLimit = cliquiz.NewQuestion(
			"目標氣體限制 Gas Limit",
			&targetGasLimit,
			cliquiz.DefaultAnswer(defautTargetGasLimit),
			cliquiz.Required(),
		)
		questionSourceAddressPrefix = cliquiz.NewQuestion(
			"源地址前綴",
			&sourceAddressPrefix,
			cliquiz.DefaultAnswer(defautSourceAddressPrefix),
			cliquiz.Required(),
		)
		questionTargetAddressPrefix = cliquiz.NewQuestion(
			"目標地址前綴",
			&targetAddressPrefix,
			cliquiz.DefaultAnswer(defautTargetAddressPrefix),
			cliquiz.Required(),
		)
	)

	// Get flags
	advanced, err := cmd.Flags().GetBool(flagAdvanced)
	if err != nil {
		return err
	}
	sourceAccount, err = cmd.Flags().GetString(flagSourceAccount)
	if err != nil {
		return err
	}
	targetAccount, err = cmd.Flags().GetString(flagTargetAccount)
	if err != nil {
		return err
	}
	sourceRPCAddress, err = cmd.Flags().GetString(flagSourceRPC)
	if err != nil {
		return err
	}
	sourceFaucetAddress, err = cmd.Flags().GetString(flagSourceFaucet)
	if err != nil {
		return err
	}
	targetRPCAddress, err = cmd.Flags().GetString(flagTargetRPC)
	if err != nil {
		return err
	}
	targetFaucetAddress, err = cmd.Flags().GetString(flagTargetFaucet)
	if err != nil {
		return err
	}
	sourcePort, err = cmd.Flags().GetString(flagSourcePort)
	if err != nil {
		return err
	}
	sourceVersion, err = cmd.Flags().GetString(flagSourceVersion)
	if err != nil {
		return err
	}
	targetPort, err = cmd.Flags().GetString(flagTargetPort)
	if err != nil {
		return err
	}
	targetVersion, err = cmd.Flags().GetString(flagTargetVersion)
	if err != nil {
		return err
	}
	sourceGasPrice, err = cmd.Flags().GetString(flagSourceGasPrice)
	if err != nil {
		return err
	}
	targetGasPrice, err = cmd.Flags().GetString(flagTargetGasPrice)
	if err != nil {
		return err
	}
	sourceGasLimit, err = cmd.Flags().GetInt64(flagSourceGasLimit)
	if err != nil {
		return err
	}
	targetGasLimit, err = cmd.Flags().GetInt64(flagTargetGasLimit)
	if err != nil {
		return err
	}
	sourceAddressPrefix, err = cmd.Flags().GetString(flagSourceAddressPrefix)
	if err != nil {
		return err
	}
	targetAddressPrefix, err = cmd.Flags().GetString(flagTargetAddressPrefix)
	if err != nil {
		return err
	}
	ordered, err := cmd.Flags().GetBool(flagOrdered)
	if err != nil {
		return err
	}
	var (
		sourceClientID, _ = cmd.Flags().GetString(flagSourceClientID)
		targetClientID, _ = cmd.Flags().GetString(flagTargetClientID)
		reset, _          = cmd.Flags().GetBool(flagReset)

		questions []cliquiz.Question
	)

	// 如果未提供標誌，則從提示中獲取信息
	if sourceAccount == "" {
		questions = append(questions, questionSourceAccount)
	}
	if targetAccount == "" {
		questions = append(questions, questionTargetAccount)
	}
	if sourceRPCAddress == "" {
		questions = append(questions, questionSourceRPCAddress)
	}
	if sourceFaucetAddress == "" {
		questions = append(questions, questionSourceFaucet)
	}
	if targetRPCAddress == "" {
		questions = append(questions, questionTargetRPCAddress)
	}
	if targetFaucetAddress == "" {
		questions = append(questions, questionTargetFaucet)
	}
	if sourceGasPrice == "" {
		questions = append(questions, questionSourceGasPrice)
	}
	if targetGasPrice == "" {
		questions = append(questions, questionTargetGasPrice)
	}
	if sourceGasLimit == 0 {
		questions = append(questions, questionSourceGasLimit)
	}
	if targetGasLimit == 0 {
		questions = append(questions, questionTargetGasLimit)
	}
	if sourceAddressPrefix == "" {
		questions = append(questions, questionSourceAddressPrefix)
	}
	if targetAddressPrefix == "" {
		questions = append(questions, questionTargetAddressPrefix)
	}
	// advanced information
	if advanced {
		if sourcePort == "" {
			questions = append(questions, questionSourcePort)
		}
		if sourceVersion == "" {
			questions = append(questions, questionSourceVersion)
		}
		if targetPort == "" {
			questions = append(questions, questionTargetPort)
		}
		if targetVersion == "" {
			questions = append(questions, questionTargetVersion)
		}
	}

	session.PauseSpinner()
	if len(questions) > 0 {
		if err := session.Ask(questions...); err != nil {
			return err
		}
	}

	if reset {
		if err := relayerconfig.Delete(); err != nil {
			return err
		}
	}

	session.StartSpinner("獲取鏈信息...")

	session.Println()
	r := relayer.New(ca)

	// 初始化鏈
	sourceChain, err := initChain(
		cmd,
		r,
		session,
		relayerSource,
		sourceAccount,
		sourceRPCAddress,
		sourceFaucetAddress,
		sourceGasPrice,
		sourceGasLimit,
		sourceAddressPrefix,
		sourceClientID,
	)
	if err != nil {
		return err
	}

	targetChain, err := initChain(
		cmd,
		r,
		session,
		relayerTarget,
		targetAccount,
		targetRPCAddress,
		targetFaucetAddress,
		targetGasPrice,
		targetGasLimit,
		targetAddressPrefix,
		targetClientID,
	)
	if err != nil {
		return err
	}

	session.StartSpinner("配置中...")

	// 設置高級頻道選項
	var channelOptions []relayer.ChannelOption
	if advanced {
		channelOptions = append(channelOptions,
			relayer.SourcePort(sourcePort),
			relayer.SourceVersion(sourceVersion),
			relayer.TargetPort(targetPort),
			relayer.TargetVersion(targetVersion),
		)

		if ordered {
			channelOptions = append(channelOptions, relayer.Ordered())
		}
	}

	// 創建連接配置
	id, err := sourceChain.Connect(targetChain, channelOptions...)
	if err != nil {
		return err
	}

	session.StopSpinner()
	session.Printf("⛓  配置的鏈: %s\n\n", color.Green.Sprint(id))

	return nil
}

// initChain 初始化中繼連接的鏈信息
func initChain(
	cmd *cobra.Command,
	r relayer.Relayer,
	session cliui.Session,
	name,
	accountName,
	rpcAddr,
	faucetAddr,
	gasPrice string,
	gasLimit int64,
	addressPrefix,
	clientID string,
) (*relayer.Chain, error) {
	defer session.StopSpinner()
	session.StartSpinner("初始化鏈...")

	c, account, err := r.NewChain(
		cmd.Context(),
		accountName,
		rpcAddr,
		relayer.WithFaucet(faucetAddr),
		relayer.WithGasPrice(gasPrice),
		relayer.WithGasLimit(gasLimit),
		relayer.WithAddressPrefix(addressPrefix),
		relayer.WithClientID(clientID),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "無法解決 %s", name)
	}

	session.StopSpinner()

	accountAddr := account.Address(addressPrefix)

	session.Printf("🔐  帳戶 %q 是 %s(%s)\n \n", name, accountName, accountAddr)
	session.StartSpinner(color.Yellow.Sprintf("試圖從水龍頭接收令牌..."))

	coins, err := c.TryRetrieve(cmd.Context())
	session.StopSpinner()

	session.Print(" |· ")
	if err != nil {
		session.Println(color.Yellow.Sprintf(err.Error()))
	} else {
		session.Println(color.Green.Sprintf("從水龍頭收到硬幣"))
	}

	balance := coins.String()
	if balance == "" {
		balance = entrywriter.None
	}
	session.Printf(" |· (平衡: %s)\n\n", balance)

	return c, nil
}
