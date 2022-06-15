package networktypes

import (
	spntypes "github.com/tendermint/spn/pkg/types"
)

//獎勵是節點及獎勵信息。
type Reward struct {
	ConsensusState spntypes.ConsensusState
	ValidatorSet   spntypes.ValidatorSet
	RevisionHeight uint64
}
