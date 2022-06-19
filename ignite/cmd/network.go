package ignitecmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/ignite-hq/cli/ignite/pkg/cosmosaccount"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosclient"
	"github.com/ignite-hq/cli/ignite/pkg/events"
	"github.com/ignite-hq/cli/ignite/pkg/gitpod"
	"github.com/ignite-hq/cli/ignite/services/network"
	"github.com/ignite-hq/cli/ignite/services/network/networkchain"
	"github.com/ignite-hq/cli/ignite/services/network/networktypes"
)

var (
	nightly bool
	local   bool

	spnNodeAddress   string
	spnFaucetAddress string
)

const (
	flagNightly = "nightly"
	flagLocal   = "local"

	flagSPNNodeAddress   = "spn-node-address"
	flagSPNFaucetAddress = "spn-faucet-address"

	spnNodeAddressNightly   = "https://rpc.nightly.starport.network:443"
	spnFaucetAddressNightly = "https://faucet.nightly.starport.network"

	spnNodeAddressLocal   = "http://0.0.0.0:26657"
	spnFaucetAddressLocal = "http://0.0.0.0:4500"
)

// NewNetwork 創建一個包含其他一些子命令的新網絡命令
// 與協作創建新網絡有關。
func NewNetwork() *cobra.Command {
	c := &cobra.Command{
		Use:     "network [command]",
		Aliases: []string{"n"},
		Short:   "在生產中啟動區塊鏈網絡",
		Args:    cobra.ExactArgs(1),
		Hidden:  true,
	}

	// configure flags.
	c.PersistentFlags().BoolVar(&local, flagLocal, false, "使用本地 SPN 網絡")
	c.PersistentFlags().BoolVar(&nightly, flagNightly, false, "使用夜間 SPN 網絡")
	c.PersistentFlags().StringVar(&spnNodeAddress, flagSPNNodeAddress, spnNodeAddressNightly, "SPN 節點地址")
	c.PersistentFlags().StringVar(&spnFaucetAddress, flagSPNFaucetAddress, spnFaucetAddressNightly, "SPN水龍頭地址")

	// add sub commands.
	c.AddCommand(
		NewNetworkChain(),
		NewNetworkCampaign(),
		NewNetworkRequest(),
		NewNetworkReward(),
		NewNetworkClient(),
	)

	return c
}

var cosmos *cosmosclient.Client

type (
	NetworkBuilderOption func(builder *NetworkBuilder)

	NetworkBuilder struct {
		AccountRegistry cosmosaccount.Registry

		ev  events.Bus
		cmd *cobra.Command
		cc  cosmosclient.Client
	}
)

func CollectEvents(ev events.Bus) NetworkBuilderOption {
	return func(builder *NetworkBuilder) {
		builder.ev = ev
	}
}

func flagSetSPNAccountPrefixes() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(flagAddressPrefix, networktypes.SPN, "賬戶地址前綴")
	return fs
}

func newNetworkBuilder(cmd *cobra.Command, options ...NetworkBuilderOption) (NetworkBuilder, error) {
	var (
		err error
		n   = NetworkBuilder{cmd: cmd}
	)

	if n.cc, err = getNetworkCosmosClient(cmd); err != nil {
		return NetworkBuilder{}, err
	}

	n.AccountRegistry = n.cc.AccountRegistry

	for _, apply := range options {
		apply(&n)
	}
	return n, nil
}

func (n NetworkBuilder) Chain(source networkchain.SourceOption, options ...networkchain.Option) (*networkchain.Chain, error) {
	if home := getHome(n.cmd); home != "" {
		options = append(options, networkchain.WithHome(home))
	}

	options = append(options, networkchain.CollectEvents(n.ev))

	return networkchain.New(n.cmd.Context(), n.AccountRegistry, source, options...)
}

func (n NetworkBuilder) Network(options ...network.Option) (network.Network, error) {
	var (
		err     error
		from    = getFrom(n.cmd)
		account = cosmosaccount.Account{}
	)
	if from != "" {
		account, err = cosmos.AccountRegistry.GetByName(getFrom(n.cmd))
		if err != nil {
			return network.Network{}, errors.Wrap(err, "確保此帳戶存在，使用 'ignite account -h' 管理帳戶")
		}
	}

	options = append(options, network.CollectEvents(n.ev))

	return network.New(*cosmos, account, options...), nil
}

func getNetworkCosmosClient(cmd *cobra.Command) (cosmosclient.Client, error) {
	// 檢查預配置的網絡
	if nightly && local {
		return cosmosclient.Client{}, errors.New("local 和 nightly 網絡不能在同一個命令中同時指定，請指定 local 或 nightly")
	}
	if local {
		spnNodeAddress = spnNodeAddressLocal
		spnFaucetAddress = spnFaucetAddressLocal
	} else if nightly {
		spnNodeAddress = spnNodeAddressNightly
		spnFaucetAddress = spnFaucetAddressNightly
	}

	cosmosOptions := []cosmosclient.Option{
		cosmosclient.WithHome(cosmosaccount.KeyringHome),
		cosmosclient.WithNodeAddress(spnNodeAddress),
		cosmosclient.WithAddressPrefix(networktypes.SPN),
		cosmosclient.WithUseFaucet(spnFaucetAddress, networktypes.SPNDenom, 5),
		cosmosclient.WithKeyringServiceName(cosmosaccount.KeyringServiceName),
	}

	keyringBackend := getKeyringBackend(cmd)
	// 在 Gitpod 上使用測試密鑰環後端，以防止提示輸入密鑰環密碼。
	// 這是因為 Gitpod 使用容器。
	// 當不在 Gitpod 上時，使用操作系統密鑰環後端，它只詢問一次密碼。
	if gitpod.IsOnGitpod() {
		keyringBackend = cosmosaccount.KeyringTest
	}
	if keyringBackend != "" {
		cosmosOptions = append(cosmosOptions, cosmosclient.WithKeyringBackend(keyringBackend))
	}

	// 在啟動時只初始化一次 cosmos 客戶端，以便spnclient在以下步驟中重用未鎖定的密鑰環。

	if cosmos == nil {
		client, err := cosmosclient.New(cmd.Context(), cosmosOptions...)
		if err != nil {
			return cosmosclient.Client{}, err
		}
		cosmos = &client
	}

	if err := cosmos.AccountRegistry.EnsureDefaultAccount(); err != nil {
		return cosmosclient.Client{}, err
	}

	return *cosmos, nil
}
