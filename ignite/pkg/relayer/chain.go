package relayer

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/imdario/mergo"

	"github.com/ignite-hq/cli/ignite/pkg/cosmosaccount"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosclient"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosfaucet"
	relayerconfig "github.com/ignite-hq/cli/ignite/pkg/relayer/config"
)

const (
	TransferPort      = "transfer"
	TransferVersion   = "ics20-1"
	OrderingUnordered = "ORDER_UNORDERED"
	OrderingOrdered   = "ORDER_ORDERED"
)

var (
	errEndpointExistsWithDifferentChainID = errors.New("Rpc端點已存在不同的Chain Id")
)

// Chain 表示中繼器中的一條鏈。
type Chain struct {
	// ID 是鏈的 ID。
	ID string

	// accountName 是鏈上使用的賬戶。
	accountName string

	// rpcAddress 是 tm 的節點地址。
	rpcAddress string

	// faucetAddress 是獲取中繼者帳戶令牌的水龍頭地址。
	faucetAddress string

	// gasPrice是向鏈發送交易時使用的 gas 價格
	gasPrice string

	// gasLimit是向鏈發送交易時使用的氣體限制
	gasLimit int64

	// addressPrefix是鏈的地址前綴。
	addressPrefix string

	// clientID是中繼器連接的鏈的客戶端 ID。
	clientID string

	r Relayer
}

// Account 代表中繼器中的一個帳戶。
type Account struct {
	// Address of the account.
	Address string `json:"address"`
}

// Option 用於配置鏈。
type Option func(*Chain)

// WithFaucet 為鏈提供了一個水龍頭地址來獲取令牌。
// 當它沒有提供時。
func WithFaucet(address string) Option {
	return func(c *Chain) {
		c.faucetAddress = address
	}
}

// WithGasPrice 給出用於將 ibc 交易發送到鏈的 gas 價格。
func WithGasPrice(gasPrice string) Option {
	return func(c *Chain) {
		c.gasPrice = gasPrice
	}
}

// WithGasLimit 給出用於將 ibc 交易發送到鏈的氣體限制。
func WithGasLimit(limit int64) Option {
	return func(c *Chain) {
		c.gasLimit = limit
	}
}

// WithAddressPrefix 配置鏈上使用的帳戶密鑰前綴。
func WithAddressPrefix(addressPrefix string) Option {
	return func(c *Chain) {
		c.addressPrefix = addressPrefix
	}
}

// WithClientID 配置鏈client id
func WithClientID(clientID string) Option {
	return func(c *Chain) {
		c.clientID = clientID
	}
}

// NewChain 在中繼器上創建新鍊或使用現有匹配鏈。
func (r Relayer) NewChain(ctx context.Context, accountName, rpcAddress string, options ...Option) (
	*Chain, cosmosaccount.Account, error) {
	c := &Chain{
		accountName: accountName,
		rpcAddress:  fixRPCAddress(rpcAddress),
		r:           r,
	}

	// 應用用戶選項。
	for _, o := range options {
		o(c)
	}

	if err := c.ensureChainSetup(ctx); err != nil {
		return nil, cosmosaccount.Account{}, err
	}

	account, err := r.ca.GetByName(accountName)
	if err != nil {
		return nil, cosmosaccount.Account{}, err
	}

	return c, account, nil
}

// TryRetrieve嘗試將一些硬幣接收到帳戶並返​​回總餘額。
func (c *Chain) TryRetrieve(ctx context.Context) (sdk.Coins, error) {
	acc, err := c.r.ca.GetByName(c.accountName)
	if err != nil {
		return nil, err
	}

	addr := acc.Address(c.addressPrefix)

	if err = cosmosfaucet.TryRetrieve(ctx, c.ID, c.rpcAddress, c.faucetAddress, addr); err != nil {
		return nil, err
	}
	return c.r.balance(ctx, c.rpcAddress, c.accountName, c.addressPrefix)
}

// channelOptions表示在兩條鏈之間配置 IBC 通道的選項
type channelOptions struct {
	sourcePort    string
	sourceVersion string
	targetPort    string
	targetVersion string
	ordering      string
}

// newChannelOptions返回默認通道選項
func newChannelOptions() channelOptions {
	return channelOptions{
		sourcePort:    TransferPort,
		sourceVersion: TransferVersion,
		targetPort:    TransferPort,
		targetVersion: TransferVersion,
		ordering:      OrderingUnordered,
	}
}

// ChannelOption用於配置中繼器 IBC 連接
type ChannelOption func(*channelOptions)

// SourcePort配置新通道的源端口
func SourcePort(port string) ChannelOption {
	return func(c *channelOptions) {
		c.sourcePort = port
	}
}

// TargetPort配置新通道的目標端口
func TargetPort(port string) ChannelOption {
	return func(c *channelOptions) {
		c.targetPort = port
	}
}

// SourceVersion 配置新通道的源版本
func SourceVersion(version string) ChannelOption {
	return func(c *channelOptions) {
		c.sourceVersion = version
	}
}

// TargetVersion配置新通道的目標版本
func TargetVersion(version string) ChannelOption {
	return func(c *channelOptions) {
		c.targetVersion = version
	}
}

// Ordered按順序設置新頻道
func Ordered() ChannelOption {
	return func(c *channelOptions) {
		c.ordering = OrderingOrdered
	}
}

// Connect 將 dst 鏈連接到 c 鏈，並在離線模式下創建一條路徑。
// 成功時返迴路徑 id，否則返回非零錯誤。
func (c *Chain) Connect(dst *Chain, options ...ChannelOption) (id string, err error) {
	channelOptions := newChannelOptions()

	for _, apply := range options {
		apply(&channelOptions)
	}

	conf, err := relayerconfig.Get()
	if err != nil {
		return "", err
	}

	//從帶有遞增數字的鏈 id 中確定唯一的路徑名。例如。：
	// - src-dst
	// - src-dst-2
	pathID := fmt.Sprintf("%s-%s", c.ID, dst.ID)
	var suffix string
	i := 2
	for {
		guess := pathID + suffix
		if _, err := conf.PathByID(guess); err != nil { //猜測是獨一無二的。
			pathID = guess
			break
		}
		suffix = fmt.Sprintf("-%d", i)
		i++
	}

	confPath := relayerconfig.Path{
		ID:       pathID,
		Ordering: channelOptions.ordering,
		Src: relayerconfig.PathEnd{
			ChainID: c.ID,
			PortID:  channelOptions.sourcePort,
			Version: channelOptions.sourceVersion,
		},
		Dst: relayerconfig.PathEnd{
			ChainID: dst.ID,
			PortID:  channelOptions.targetPort,
			Version: channelOptions.targetVersion,
		},
	}

	conf.Paths = append(conf.Paths, confPath)

	if err := relayerconfig.Save(conf); err != nil {
		return "", err
	}

	return pathID, nil
}

// ensureChainSetup建立新的或現有的鏈。
func (c *Chain) ensureChainSetup(ctx context.Context) error {
	client, err := cosmosclient.New(ctx, cosmosclient.WithNodeAddress(c.rpcAddress))
	if err != nil {
		return err
	}
	status, err := client.RPC.Status(ctx)
	if err != nil {
		return err
	}
	c.ID = status.NodeInfo.Network

	confChain := relayerconfig.Chain{
		ID:            c.ID,
		Account:       c.accountName,
		AddressPrefix: c.addressPrefix,
		RPCAddress:    c.rpcAddress,
		GasPrice:      c.gasPrice,
		GasLimit:      c.gasLimit,
		ClientID:      c.clientID,
	}

	conf, err := relayerconfig.Get()
	if err != nil {
		return err
	}

	var found bool

	for i, chain := range conf.Chains {
		if chain.ID == c.ID {
			if chain.RPCAddress != c.rpcAddress {
				return errEndpointExistsWithDifferentChainID
			}

			if err := mergo.Merge(&conf.Chains[i], confChain, mergo.WithOverride); err != nil {
				return err
			}

			found = true
			break
		}
	}

	if !found {
		conf.Chains = append(conf.Chains, confChain)
	}

	return relayerconfig.Save(conf)
}
