package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v2/x/oracle/types"
)

// GetVoteTargets returns the vote target list
func (k Keeper) GetVoteTargets(ctx sdk.Context) ([]string, error) {
	var voteTargets []string
	err := k.VoteTarget.Walk(ctx, nil, func(denom string, denomInfo types.Denom) (bool, error) {
		voteTargets = append(voteTargets, denom)
		return false, nil
	})
	return voteTargets, err
}
