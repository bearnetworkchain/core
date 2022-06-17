package network

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	rewardtypes "github.com/tendermint/spn/x/reward/types"

	"github.com/bearnetworkchain/core/ignite/pkg/cliui/icons"
	"github.com/bearnetworkchain/core/ignite/pkg/events"
	"github.com/bearnetworkchain/core/ignite/services/network/networktypes"
)

// SetReward設置連鎖獎勵
func (n Network) SetReward(launchID uint64, lastRewardHeight int64, coins sdk.Coins) error {
	n.ev.Send(events.New(
		events.StatusOngoing,
		fmt.Sprintf("設置獎勵 %s 到鏈 %d 在高度 %d",
			coins.String(),
			launchID,
			lastRewardHeight,
		),
	))

	msg := rewardtypes.NewMsgSetRewards(
		n.account.Address(networktypes.SPN),
		launchID,
		lastRewardHeight,
		coins,
	)
	res, err := n.cosmos.BroadcastTx(n.account.Name, msg)
	if err != nil {
		return err
	}

	var setRewardRes rewardtypes.MsgSetRewardsResponse
	if err := res.Decode(&setRewardRes); err != nil {
		return err
	}

	if setRewardRes.PreviousCoins.Empty() {
		n.ev.Send(events.New(
			events.StatusDone,
			"獎勵池為空",
			events.Icon(icons.Info),
		))
	} else {
		n.ev.Send(events.New(events.StatusDone,
			fmt.Sprintf(
				"以前的獎勵池 %s 在高度 %d 被覆蓋",
				coins.String(),
				lastRewardHeight,
			),
			events.Icon(icons.Info),
		))
	}

	if setRewardRes.NewCoins.Empty() {
		n.ev.Send(events.New(events.StatusDone, "獎勵池被移除"))
	} else {
		n.ev.Send(events.New(events.StatusDone, fmt.Sprintf(
			"%s 將分發給高度的驗證者 %d. 這條鏈 %d 現在是一個激勵測試網",
			coins.String(),
			lastRewardHeight,
			launchID,
		)))
	}
	return nil
}
