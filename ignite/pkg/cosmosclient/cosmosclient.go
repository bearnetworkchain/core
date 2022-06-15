// Package cosmosclient 提供一個獨立的客戶端來連接 Cosmos SDK 鏈。
package cosmosclient

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	commitmenttypes "github.com/cosmos/ibc-go/v3/modules/core/23-commitment/types"
	"github.com/gogo/protobuf/proto"
	prototypes "github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/bytes"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/ignite-hq/cli/ignite/pkg/cosmosaccount"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosfaucet"
)

// FaucetTransferEnsureDuration 是當水龍頭傳輸時 BroadcastTx 將等待的持續時間
// 在廣播之前觸發，但傳輸的 tx 尚未在狀態下提交。
var FaucetTransferEnsureDuration = time.Second * 40

var errCannotRetrieveFundsFromFaucet = errors.New("無法從水龍頭中取回資金")

const (
	defaultNodeAddress   = "http://localhost:26657"
	defaultGasAdjustment = 1.0
	defaultGasLimit      = 300000
)

const (
	defaultFaucetAddress   = "http://localhost:4500"
	defaultFaucetDenom     = "token"
	defaultFaucetMinAmount = 100
)

// 客戶端是通過查詢和廣播交易來訪問您的鏈的客戶端。
type Client struct {
	// RPC 是 Tendermint RPC。
	RPC *rpchttp.HTTP

	// Factory 是 Cosmos SDK tx 工廠。
	Factory tx.Factory

	// context 是 Cosmos SDK 客戶端上下文。
	context client.Context

	// AccountRegistry 是訪問帳戶的重試。
	AccountRegistry cosmosaccount.Registry

	addressPrefix string

	nodeAddress string
	out         io.Writer
	chainID     string

	useFaucet       bool
	faucetAddress   string
	faucetDenom     string
	faucetMinAmount uint64

	homePath           string
	keyringServiceName string
	keyringBackend     cosmosaccount.KeyringBackend
}

// 選項配置您的客戶端。
type Option func(*Client)

// WithHome 設置鏈的數據目錄。此選項用於訪問您的鏈
// 基於文件的密鑰環，僅在您處理創建和簽署交易時才需要。
// 如果沒有提供，您的數據目錄將被假定為 `$HOME/.your-chain-id`.
func WithHome(path string) Option {
	return func(c *Client) {
		c.homePath = path
	}
}

// 當您使用操作系統密鑰環後端時，WithKeyringServiceName 用作密鑰環的名稱。
// 默認情況下是 `cosmos`.
func WithKeyringServiceName(name string) Option {
	return func(c *Client) {
		c.keyringServiceName = name
	}
}

// WithKeyringBackend 設置您的密鑰環後端。默認情況下，它是 `test`.
func WithKeyringBackend(backend cosmosaccount.KeyringBackend) Option {
	return func(c *Client) {
		c.keyringBackend = backend
	}
}

// WithNodeAddress 設置鏈的節點地址。未提供此選項時
// `http://localhost:26657` 用作默認值。
func WithNodeAddress(addr string) Option {
	return func(c *Client) {
		c.nodeAddress = addr
	}
}

func WithAddressPrefix(prefix string) Option {
	return func(c *Client) {
		c.addressPrefix = prefix
	}
}

func WithUseFaucet(faucetAddress, denom string, minAmount uint64) Option {
	return func(c *Client) {
		c.useFaucet = true
		c.faucetAddress = faucetAddress
		if denom != "" {
			c.faucetDenom = denom
		}
		if minAmount != 0 {
			c.faucetMinAmount = minAmount
		}
	}
}

// New 創建具有給定選項的新客戶端。
func New(ctx context.Context, options ...Option) (Client, error) {
	c := Client{
		nodeAddress:     defaultNodeAddress,
		keyringBackend:  cosmosaccount.KeyringTest,
		addressPrefix:   "bnkt",
		faucetAddress:   defaultFaucetAddress,
		faucetDenom:     defaultFaucetDenom,
		faucetMinAmount: defaultFaucetMinAmount,
		out:             io.Discard,
	}

	var err error

	for _, apply := range options {
		apply(&c)
	}

	if c.RPC, err = rpchttp.New(c.nodeAddress, "/websocket"); err != nil {
		return Client{}, err
	}

	statusResp, err := c.RPC.Status(ctx)
	if err != nil {
		return Client{}, err
	}

	c.chainID = statusResp.NodeInfo.Network

	if c.homePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Client{}, err
		}
		c.homePath = filepath.Join(home, "."+c.chainID)
	}

	c.AccountRegistry, err = cosmosaccount.New(
		cosmosaccount.WithKeyringServiceName(c.keyringServiceName),
		cosmosaccount.WithKeyringBackend(c.keyringBackend),
		cosmosaccount.WithHome(c.homePath),
	)
	if err != nil {
		return Client{}, err
	}

	c.context = newContext(c.RPC, c.out, c.chainID, c.homePath).WithKeyring(c.AccountRegistry.Keyring)
	c.Factory = newFactory(c.context)

	return c, nil
}

func (c Client) Account(accountName string) (cosmosaccount.Account, error) {
	return c.AccountRegistry.GetByName(accountName)
}

// Address 從賬戶名返回賬戶地址.
func (c Client) Address(accountName string) (sdktypes.AccAddress, error) {
	account, err := c.Account(accountName)
	if err != nil {
		return sdktypes.AccAddress{}, err
	}
	return account.Info.GetAddress(), nil
}

func (c Client) Context() client.Context {
	return c.context
}

// Response 您的廣播交易.
type Response struct {
	Codec codec.Codec

	// TxResponse 是底層 tx 響應.
	*sdktypes.TxResponse
}

// Decode 將您的 Msg 服務中定義的 proto func 響應解碼為您的消息類型。
// message 需要是一個指針。並且您需要為解碼函數提供正確的原始消息（結構）類型。
//
// 例如，對於以下 CreateChain 函數，類型將是: `types.MsgCreateChainResponse`.
//
// ```proto
// service Msg {
//   rpc CreateChain(MsgCreateChain) returns (MsgCreateChainResponse);
// }
// ```
func (r Response) Decode(message proto.Message) error {
	data, err := hex.DecodeString(r.Data)
	if err != nil {
		return err
	}

	var txMsgData sdktypes.TxMsgData
	if err := r.Codec.Unmarshal(data, &txMsgData); err != nil {
		return err
	}

	resData := txMsgData.Data[0]

	return prototypes.UnmarshalAny(&prototypes.Any{
		// TODO 動態獲取類型 url(basically remove `+ "Response"`) 以下問題解決後。
		// https://github.com/cosmos/cosmos-sdk/issues/10496
		TypeUrl: resData.MsgType + "Response",
		Value:   resData.Data,
	}, message)
}

// ConsensusInfo 是驗證者共識信息
type ConsensusInfo struct {
	Timestamp          string                `json:"Timestamp"`
	Root               string                `json:"Root"`
	NextValidatorsHash string                `json:"NextValidatorsHash"`
	ValidatorSet       *tmproto.ValidatorSet `json:"ValidatorSet"`
}

// ConsensusInfo 通過給定高度返回適當的tendermint 共識狀態
// 以及為下一個高度設置的驗證器
func (c Client) ConsensusInfo(ctx context.Context, height int64) (ConsensusInfo, error) {
	node, err := c.Context().GetNode()
	if err != nil {
		return ConsensusInfo{}, err
	}

	commit, err := node.Commit(ctx, &height)
	if err != nil {
		return ConsensusInfo{}, err
	}

	var (
		page  = 1
		count = 10_000
	)
	validators, err := node.Validators(ctx, &height, &page, &count)
	if err != nil {
		return ConsensusInfo{}, err
	}

	protoValset, err := tmtypes.NewValidatorSet(validators.Validators).ToProto()
	if err != nil {
		return ConsensusInfo{}, err
	}

	heightNext := height + 1
	validatorsNext, err := node.Validators(ctx, &heightNext, &page, &count)
	if err != nil {
		return ConsensusInfo{}, err
	}

	var (
		hash = tmtypes.NewValidatorSet(validatorsNext.Validators).Hash()
		root = commitmenttypes.NewMerkleRoot(commit.AppHash)
	)

	return ConsensusInfo{
		Timestamp:          commit.Time.Format(time.RFC3339Nano),
		NextValidatorsHash: bytes.HexBytes(hash).String(),
		Root:               base64.StdEncoding.EncodeToString(root.Hash),
		ValidatorSet:       protoValset,
	}, nil
}

// status 返回節點狀態
func (c Client) Status(ctx context.Context) (*ctypes.ResultStatus, error) {
	return c.RPC.Status(ctx)
}

// BroadcastTx 創建並廣播帶有給定消息的 tx 以供帳戶使用。
func (c Client) BroadcastTx(accountName string, msgs ...sdktypes.Msg) (Response, error) {
	_, broadcast, err := c.BroadcastTxWithProvision(accountName, msgs...)
	if err != nil {
		return Response{}, err
	}
	return broadcast()
}

// 保護 sdktypes.Config。
var mconf sync.Mutex

func (c Client) BroadcastTxWithProvision(accountName string, msgs ...sdktypes.Msg) (
	gas uint64, broadcast func() (Response, error), err error) {
	if err := c.prepareBroadcast(context.Background(), accountName, msgs); err != nil {
		return 0, nil, err
	}

	// TODO 如果可能的話，找到更好的方法。
	mconf.Lock()
	defer mconf.Unlock()
	config := sdktypes.GetConfig()
	config.SetBech32PrefixForAccount(c.addressPrefix, c.addressPrefix+"pub")

	accountAddress, err := c.Address(accountName)
	if err != nil {
		return 0, nil, err
	}

	ctx := c.context.
		WithFromName(accountName).
		WithFromAddress(accountAddress)

	txf, err := prepareFactory(ctx, c.Factory)
	if err != nil {
		return 0, nil, err
	}

	_, gas, err = tx.CalculateGas(ctx, txf, msgs...)
	if err != nil {
		return 0, nil, err
	}
	// 模擬氣體可能與實際交易所需的實際氣體不同
	// 我們添加額外的量以承受足夠的氣體提供
	gas += 10000
	txf = txf.WithGas(gas)

	// 返回提供函數
	return gas, func() (Response, error) {
		txUnsigned, err := tx.BuildUnsignedTx(txf, msgs...)
		if err != nil {
			return Response{}, err
		}

		txUnsigned.SetFeeGranter(ctx.GetFeeGranterAddress())
		if err := tx.Sign(txf, accountName, txUnsigned, true); err != nil {
			return Response{}, err
		}

		txBytes, err := ctx.TxConfig.TxEncoder()(txUnsigned.GetTx())
		if err != nil {
			return Response{}, err
		}

		resp, err := ctx.BroadcastTx(txBytes)
		if err == sdkerrors.ErrInsufficientFunds {
			err = c.makeSureAccountHasTokens(context.Background(), accountAddress.String())
			if err != nil {
				return Response{}, err
			}
			resp, err = ctx.BroadcastTx(txBytes)
		}

		return Response{
			Codec:      ctx.Codec,
			TxResponse: resp,
		}, handleBroadcastResult(resp, err)
	}, nil
}

// prepareBroadcast 在廣播消息之前執行檢查和操作
func (c *Client) prepareBroadcast(ctx context.Context, accountName string, _ []sdktypes.Msg) error {
	// TODO 之後取消評論 https://github.com/tendermint/spn/issues/363
	// validate msgs.
	//  for _, msg := range msgs {
	//  if err := msg.ValidateBasic(); err != nil {
	//  return err
	//  }
	//  }

	account, err := c.Account(accountName)
	if err != nil {
		return err
	}

	// 在廣播之前確保該帳戶有足夠的餘額。
	if c.useFaucet {
		if err := c.makeSureAccountHasTokens(ctx, account.Address(c.addressPrefix)); err != nil {
			return err
		}
	}

	return nil
}

// makeSureAccountHasTokens確保地址具有正餘額
// 如果地址有空餘額，它會從水龍頭請求資金
func (c *Client) makeSureAccountHasTokens(ctx context.Context, address string) error {
	if err := c.checkAccountBalance(ctx, address); err == nil {
		return nil
	}

	//從水龍頭索取硬幣。
	fc := cosmosfaucet.NewClient(c.faucetAddress)
	faucetResp, err := fc.Transfer(ctx, cosmosfaucet.TransferRequest{AccountAddress: address})
	if err != nil {
		return errors.Wrap(errCannotRetrieveFundsFromFaucet, err.Error())
	}
	if faucetResp.Error != "" {
		return errors.Wrap(errCannotRetrieveFundsFromFaucet, faucetResp.Error)
	}

	// 確保資金被收回。
	ctx, cancel := context.WithTimeout(ctx, FaucetTransferEnsureDuration)
	defer cancel()

	return backoff.Retry(func() error {
		return c.checkAccountBalance(ctx, address)
	}, backoff.WithContext(backoff.NewConstantBackOff(time.Second), ctx))
}

func (c *Client) checkAccountBalance(ctx context.Context, address string) error {
	resp, err := banktypes.NewQueryClient(c.context).Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: address,
		Denom:   c.faucetDenom,
	})
	if err != nil {
		return err
	}

	if resp.Balance.Amount.Uint64() >= c.faucetMinAmount {
		return nil
	}

	return fmt.Errorf("賬戶不夠 %q balance, min. 所需金額: %d", c.faucetDenom, c.faucetMinAmount)
}

// handleBroadcastResult 處理廣播消息結果的結果並檢查是否發生錯誤
func handleBroadcastResult(resp *sdktypes.TxResponse, err error) error {
	if err != nil {
		if strings.Contains(err.Error(), "未找到") {
			return errors.New("確保您的 SPN 帳戶有足夠的餘額")
		}

		return err
	}

	if resp.Code > 0 {
		return fmt.Errorf("SPN 錯誤與 '%d' 代碼: %s", resp.Code, resp.RawLog)
	}
	return nil
}

func prepareFactory(clientCtx client.Context, txf tx.Factory) (tx.Factory, error) {
	from := clientCtx.GetFromAddress()

	if err := txf.AccountRetriever().EnsureExists(clientCtx, from); err != nil {
		return txf, err
	}

	initNum, initSeq := txf.AccountNumber(), txf.Sequence()
	if initNum == 0 || initSeq == 0 {
		num, seq, err := txf.AccountRetriever().GetAccountNumberSequence(clientCtx, from)
		if err != nil {
			return txf, err
		}

		if initNum == 0 {
			txf = txf.WithAccountNumber(num)
		}

		if initSeq == 0 {
			txf = txf.WithSequence(seq)
		}
	}

	return txf, nil
}

func newContext(
	c *rpchttp.HTTP,
	out io.Writer,
	chainID,
	home string,
) client.Context {
	var (
		amino             = codec.NewLegacyAmino()
		interfaceRegistry = codectypes.NewInterfaceRegistry()
		marshaler         = codec.NewProtoCodec(interfaceRegistry)
		txConfig          = authtx.NewTxConfig(marshaler, authtx.DefaultSignModes)
	)

	authtypes.RegisterInterfaces(interfaceRegistry)
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	sdktypes.RegisterInterfaces(interfaceRegistry)
	staking.RegisterInterfaces(interfaceRegistry)
	cryptocodec.RegisterInterfaces(interfaceRegistry)

	return client.Context{}.
		WithChainID(chainID).
		WithInterfaceRegistry(interfaceRegistry).
		WithCodec(marshaler).
		WithTxConfig(txConfig).
		WithLegacyAmino(amino).
		WithInput(os.Stdin).
		WithOutput(out).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(home).
		WithClient(c).
		WithSkipConfirmation(true)
}

func newFactory(clientCtx client.Context) tx.Factory {
	return tx.Factory{}.
		WithChainID(clientCtx.ChainID).
		WithKeybase(clientCtx.Keyring).
		WithGas(defaultGasLimit).
		WithGasAdjustment(defaultGasAdjustment).
		WithSignMode(signing.SignMode_SIGN_MODE_UNSPECIFIED).
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithTxConfig(clientCtx.TxConfig)
}
