package network

import (
	"context"
	"fmt"
	"time"

	launchtypes "github.com/tendermint/spn/x/launch/types"

	"github.com/bearnetworkchain/core/ignite/pkg/events"
	"github.com/bearnetworkchain/core/ignite/pkg/xtime"
	"github.com/bearnetworkchain/core/ignite/services/network/networktypes"
)

// LaunchParams 從 SPN 獲取鏈啟動模塊參數
func (n Network) LaunchParams(ctx context.Context) (launchtypes.Params, error) {
	res, err := n.launchQuery.Params(ctx, &launchtypes.QueryParamsRequest{})
	if err != nil {
		return launchtypes.Params{}, err
	}
	return res.GetParams(), nil
}

// TriggerLaunch作為協調者啟動一個鏈
func (n Network) TriggerLaunch(ctx context.Context, launchID uint64, remainingTime time.Duration) error {
	n.ev.Send(events.New(events.StatusOngoing, fmt.Sprintf("啟動鏈 %d", launchID)))
	params, err := n.LaunchParams(ctx)
	if err != nil {
		return err
	}

	var (
		minLaunch = xtime.Seconds(params.LaunchTimeRange.MinLaunchTime)
		maxLaunch = xtime.Seconds(params.LaunchTimeRange.MaxLaunchTime)
		address   = n.account.Address(networktypes.SPN)
	)
	switch {
	case remainingTime == 0:
		// 如果用戶沒有指定剩餘時間，則使用最小的一個
		remainingTime = minLaunch
	case remainingTime < minLaunch:
		return fmt.Errorf("剩餘時間 %s 低於最小值 %s",
			xtime.NowAfter(remainingTime),
			xtime.NowAfter(minLaunch))
	case remainingTime > maxLaunch:
		return fmt.Errorf("剩餘時間 %s 大於最大值 %s",
			xtime.NowAfter(remainingTime),
			xtime.NowAfter(maxLaunch))
	}

	msg := launchtypes.NewMsgTriggerLaunch(address, launchID, int64(remainingTime.Seconds()))
	n.ev.Send(events.New(events.StatusOngoing, "設置啟動時間"))
	res, err := n.cosmos.BroadcastTx(n.account.Name, msg)
	if err != nil {
		return err
	}

	var launchRes launchtypes.MsgTriggerLaunchResponse
	if err := res.Decode(&launchRes); err != nil {
		return err
	}

	n.ev.Send(events.New(events.StatusDone,
		fmt.Sprintf("鏈 %d 將會啟動於 %s", launchID, xtime.NowAfter(remainingTime)),
	))
	return nil
}

// RevertLaunch 將已啟動的鏈恢復為協調器
func (n Network) RevertLaunch(launchID uint64, chain Chain) error {
	n.ev.Send(events.New(events.StatusOngoing, fmt.Sprintf("恢復已啟動的鏈 %d", launchID)))

	address := n.account.Address(networktypes.SPN)
	msg := launchtypes.NewMsgRevertLaunch(address, launchID)
	_, err := n.cosmos.BroadcastTx(n.account.Name, msg)
	if err != nil {
		return err
	}

	n.ev.Send(events.New(events.StatusDone,
		fmt.Sprintf("鏈 %d 啟動被恢復", launchID),
	))

	n.ev.Send(events.New(events.StatusOngoing, "重置創世時間"))
	if err := chain.ResetGenesisTime(); err != nil {
		return err
	}
	n.ev.Send(events.New(events.StatusDone, "創世時間被重置"))
	return nil
}
