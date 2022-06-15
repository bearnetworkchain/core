package network

import (
	"context"
	"encoding/base64"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	spntypes "github.com/tendermint/spn/pkg/types"

	"github.com/ignite-hq/cli/ignite/services/network/networktypes"
)

// Node是節點構建器。
type Node struct {
	cosmos       CosmosClient
	stakingQuery stakingtypes.QueryClient
}

func NewNodeClient(cosmos CosmosClient) (Node, error) {
	return Node{
		cosmos:       cosmos,
		stakingQuery: stakingtypes.NewQueryClient(cosmos.Context()),
	}, nil
}

// RewardsInfo 使用驗證器集獲取共識狀態
func RewardsInfo(ctx context.Context, client CosmosClient, height int64) (networktypes.Reward, error) {
	consensusState, err := client.ConsensusInfo(ctx, height)
	if err != nil {
		return networktypes.Reward{}, err
	}
	spnConsensusState := spntypes.NewConsensusState(
		consensusState.Timestamp,
		consensusState.NextValidatorsHash,
		consensusState.Root,
	)

	validators := make([]spntypes.Validator, len(consensusState.ValidatorSet.Validators))
	for i, validator := range consensusState.ValidatorSet.Validators {
		validators[i] = spntypes.NewValidator(
			base64.StdEncoding.EncodeToString(validator.PubKey.GetEd25519()),
			validator.ProposerPriority,
			validator.VotingPower,
		)
	}

	return networktypes.Reward{
		ConsensusState: spnConsensusState,
		ValidatorSet:   spntypes.NewValidatorSet(validators...),
		RevisionHeight: uint64(height),
	}, nil
}

// StakingParams 獲取質押模塊參數
func (n Node) StakingParams(ctx context.Context) (stakingtypes.Params, error) {
	res, err := n.stakingQuery.Params(ctx, &stakingtypes.QueryParamsRequest{})
	if err != nil {
		return stakingtypes.Params{}, err
	}
	return res.Params, nil
}

// RewardsInfo 使用驗證器集和無限制時間獲取共識狀態
func (n Node) RewardsInfo(ctx context.Context) (networktypes.Reward, int64, error) {
	status, err := n.cosmos.Status(ctx)
	if err != nil {
		return networktypes.Reward{}, 0, err
	}
	lastBlockHeight := status.SyncInfo.LatestBlockHeight

	info, err := RewardsInfo(ctx, n.cosmos, lastBlockHeight)
	if err != nil {
		return networktypes.Reward{}, 0, err
	}

	stakingParams, err := n.StakingParams(ctx)
	if err != nil {
		return networktypes.Reward{}, 0, err
	}
	return info, int64(stakingParams.UnbondingTime.Seconds()), nil
}
