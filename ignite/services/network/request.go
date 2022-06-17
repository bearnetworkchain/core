package network

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	launchtypes "github.com/tendermint/spn/x/launch/types"

	"github.com/bearnetworkchain/core/ignite/pkg/events"
	"github.com/bearnetworkchain/core/ignite/services/network/networktypes"
)

// Reviewal保留請求的審查。
type Reviewal struct {
	RequestID  uint64
	IsApproved bool
}

// ApproveRequest返回對帶有 id 的請求的批准。
func ApproveRequest(requestID uint64) Reviewal {
	return Reviewal{
		RequestID:  requestID,
		IsApproved: true,
	}
}

// RejectRequest對帶有 id 的請求返回拒絕。
func RejectRequest(requestID uint64) Reviewal {
	return Reviewal{
		RequestID:  requestID,
		IsApproved: false,
	}
}

// Requests通過啟動 ID 從 SPN 獲取所有鏈請求
func (n Network) Requests(ctx context.Context, launchID uint64) ([]networktypes.Request, error) {
	res, err := n.launchQuery.RequestAll(ctx, &launchtypes.QueryAllRequestRequest{
		LaunchID: launchID,
	})
	if err != nil {
		return nil, err
	}
	requests := make([]networktypes.Request, len(res.Request))
	for i, req := range res.Request {
		requests[i] = networktypes.ToRequest(req)
	}
	return requests, nil
}

// Request通過啟動和請求 id 從 SPN 獲取鏈請求
func (n Network) Request(ctx context.Context, launchID, requestID uint64) (networktypes.Request, error) {
	res, err := n.launchQuery.Request(ctx, &launchtypes.QueryGetRequestRequest{
		LaunchID:  launchID,
		RequestID: requestID,
	})
	if err != nil {
		return networktypes.Request{}, err
	}
	return networktypes.ToRequest(res.Request), nil
}

// RequestFromIDs 獲取通過啟動和提供的請求 ID 從 SPN 請求的鏈
// TODO：一旦實現，使用來自 https://github.com/tendermint/spn/issues/420 的 SPN 查詢
func (n Network) RequestFromIDs(ctx context.Context, launchID uint64, requestIDs ...uint64) (reqs []networktypes.Request, err error) {
	for _, id := range requestIDs {
		req, err := n.Request(ctx, launchID, id)
		if err != nil {
			return reqs, err
		}
		reqs = append(reqs, req)
	}
	return reqs, nil
}

// SubmitRequest批量提交對鏈的提案的審查。
func (n Network) SubmitRequest(launchID uint64, reviewal ...Reviewal) error {
	n.ev.Send(events.New(events.StatusOngoing, "提交請求..."))

	messages := make([]sdk.Msg, len(reviewal))
	for i, reviewal := range reviewal {
		messages[i] = launchtypes.NewMsgSettleRequest(
			n.account.Address(networktypes.SPN),
			launchID,
			reviewal.RequestID,
			reviewal.IsApproved,
		)
	}

	res, err := n.cosmos.BroadcastTx(n.account.Name, messages...)
	if err != nil {
		return err
	}

	var requestRes launchtypes.MsgSettleRequestResponse
	return res.Decode(&requestRes)
}
