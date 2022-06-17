package network

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	launchtypes "github.com/tendermint/spn/x/launch/types"

	"github.com/bearnetworkchain/core/ignite/pkg/cosmoserror"
	"github.com/bearnetworkchain/core/ignite/pkg/cosmosutil"
	"github.com/bearnetworkchain/core/ignite/pkg/events"
	"github.com/bearnetworkchain/core/ignite/pkg/xurl"
	"github.com/bearnetworkchain/core/ignite/services/network/networkchain"
	"github.com/bearnetworkchain/core/ignite/services/network/networktypes"
)

type joinOptions struct {
	accountAmount sdk.Coins
	gentxPath     string
	publicAddress string
}

type JoinOption func(*joinOptions)

func WithAccountRequest(amount sdk.Coins) JoinOption {
	return func(o *joinOptions) {
		o.accountAmount = amount
	}
}

// TODO 接受結構而不是文件路徑
func WithCustomGentxPath(path string) JoinOption {
	return func(o *joinOptions) {
		o.gentxPath = path
	}
}

func WithPublicAddress(addr string) JoinOption {
	return func(o *joinOptions) {
		o.publicAddress = addr
	}
}

// 加入網絡。
func (n Network) Join(
	ctx context.Context,
	c Chain,
	launchID uint64,
	options ...JoinOption,
) error {
	o := joinOptions{}
	for _, apply := range options {
		apply(&o)
	}

	isCustomGentx := o.gentxPath != ""
	var (
		nodeID string
		peer   launchtypes.Peer
		err    error
	)

	// 如果未提供自定義 gentx，則從鍊主文件夾獲取鏈默認值。
	if !isCustomGentx {
		if nodeID, err = c.NodeID(ctx); err != nil {
			return err
		}

		if xurl.IsHTTP(o.publicAddress) {
			peer = launchtypes.NewPeerTunnel(nodeID, networkchain.HTTPTunnelChisel, o.publicAddress)
		} else {
			peer = launchtypes.NewPeerConn(nodeID, o.publicAddress)

		}

		if o.gentxPath, err = c.DefaultGentxPath(); err != nil {
			return err
		}
	}

	// 解析 gentx 內容
	gentxInfo, gentx, err := cosmosutil.GentxFromPath(o.gentxPath)
	if err != nil {
		return err
	}

	if isCustomGentx {
		if peer, err = ParsePeerAddress(gentxInfo.Memo); err != nil {
			return err
		}
	}

	// 從主文件夾獲取鏈創世路徑
	genesisPath, err := c.GenesisPath()
	if err != nil {
		return err
	}

	// 將鏈地址前綴更改為 spn
	accountAddress, err := cosmosutil.ChangeAddressPrefix(gentxInfo.DelegatorAddress, networktypes.SPN)
	if err != nil {
		return err
	}

	if !o.accountAmount.IsZero() {
		if err := n.ensureAccount(
			ctx,
			genesisPath,
			isCustomGentx,
			launchID,
			accountAddress,
			o.accountAmount,
		); err != nil {
			return err
		}
	}

	return n.sendValidatorRequest(ctx, launchID, peer, accountAddress, gentx, gentxInfo)
}

// ensureAccount 創建添加 AddAccount 請求消息。
func (n Network) ensureAccount(
	ctx context.Context,
	genesisPath string,
	isCustomGentx bool,
	launchID uint64,
	address string,
	amount sdk.Coins,
) (err error) {
	n.ev.Send(events.New(events.StatusOngoing, "驗證帳戶已存在 "+address))

	// if is custom gentx path, avoid to check account into genesis from the home folder
	var accExist bool
	if !isCustomGentx {
		accExist, err = cosmosutil.CheckGenesisContainsAddress(genesisPath, address)
		if err != nil {
			return err
		}
		if accExist {
			return fmt.Errorf("帳戶 %s 已經存在", address)
		}
	}
	// 檢查帳戶是否作為創世帳戶存在於 SPN 鏈啟動信息中
	hasAccount, err := n.hasAccount(ctx, launchID, address)
	if err != nil {
		return err
	}
	if hasAccount {
		return fmt.Errorf("帳戶 %s 已經存在", address)
	}

	return n.sendAccountRequest(launchID, address, amount)
}

// sendValidatorRequest 在 SPN 中創建 RequestAddValidator 消息
func (n Network) sendValidatorRequest(
	ctx context.Context,
	launchID uint64,
	peer launchtypes.Peer,
	valAddress string,
	gentx []byte,
	gentxInfo cosmosutil.GentxInfo,
) error {
	// 檢查驗證器請求是否已經存在
	hasValidator, err := n.hasValidator(ctx, launchID, valAddress)
	if err != nil {
		return err
	}
	if hasValidator {
		return fmt.Errorf("驗證器 %s 已經存在", valAddress)
	}

	msg := launchtypes.NewMsgRequestAddValidator(
		n.account.Address(networktypes.SPN),
		launchID,
		valAddress,
		gentx,
		gentxInfo.PubKey,
		gentxInfo.SelfDelegation,
		peer,
	)

	n.ev.Send(events.New(events.StatusOngoing, "廣播驗證者交易"))

	res, err := n.cosmos.BroadcastTx(n.account.Name, msg)
	if err != nil {
		return err
	}

	var requestRes launchtypes.MsgRequestAddValidatorResponse
	if err := res.Decode(&requestRes); err != nil {
		return err
	}

	if requestRes.AutoApproved {
		n.ev.Send(events.New(events.StatusDone, "由協調者添加到網絡的驗證者!"))
	} else {
		n.ev.Send(events.New(events.StatusDone,
			fmt.Sprintf("要求 %d 已提交作為驗證人加入網絡!",
				requestRes.RequestID),
		))
	}
	return nil
}

// hasValidator驗證驗證器是否已存在於 SPN 存儲中
func (n Network) hasValidator(ctx context.Context, launchID uint64, address string) (bool, error) {
	_, err := n.launchQuery.GenesisValidator(ctx, &launchtypes.QueryGetGenesisValidatorRequest{
		LaunchID: launchID,
		Address:  address,
	})
	if cosmoserror.Unwrap(err) == cosmoserror.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// hasAccount 驗證帳戶是否已存在於 SPN 存儲中
func (n Network) hasAccount(ctx context.Context, launchID uint64, address string) (bool, error) {
	_, err := n.launchQuery.VestingAccount(ctx, &launchtypes.QueryGetVestingAccountRequest{
		LaunchID: launchID,
		Address:  address,
	})
	if cosmoserror.Unwrap(err) == cosmoserror.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	_, err = n.launchQuery.GenesisAccount(ctx, &launchtypes.QueryGetGenesisAccountRequest{
		LaunchID: launchID,
		Address:  address,
	})
	if cosmoserror.Unwrap(err) == cosmoserror.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
