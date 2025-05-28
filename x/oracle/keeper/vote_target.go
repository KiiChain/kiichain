package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
)

// IsVoteTarget returns true or false if the input denom is on the Vote target list
func (k Keeper) IsVoteTarget(ctx sdk.Context, denom string) bool {
	_, err := k.GetVoteTarget(ctx, denom)
	return err == nil
}

// GetVoteTargets returns the vote target list
func (k Keeper) GetVoteTargets(ctx sdk.Context) []string {
	var voteTargets []string
	k.IterateVoteTargets(ctx, func(denom string, denomInfo types.Denom) (bool, error) {
		voteTargets = append(voteTargets, denom)
		return false, nil
	})
	return voteTargets
}
