package keeper

import (
	"strconv"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v1/x/oracle/types"
)

// SlashAndResetCounters calculate if the validator must be slashed if success votes / total votes
// is lower than MinValidPerWindow param. Then reset the vote penalty info
func (k Keeper) SlashAndResetCounters(ctx sdk.Context) error {
	height := ctx.BlockHeight()
	distributionHeight := height - sdk.ValidatorUpdateDelay - 1

	// Get the module params
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	minValidPerWindow := params.MinValidPerWindow
	slashFraction := params.SlashFraction
	powerReduction := k.StakingKeeper.PowerReduction(ctx)

	// Iterate each voting result per validator
	err = k.VotePenaltyCounter.Walk(ctx, nil, func(operator sdk.ValAddress, votePenaltyCounter types.VotePenaltyCounter) (bool, error) {
		successCount := votePenaltyCounter.SuccessCount
		abstainCount := votePenaltyCounter.AbstainCount
		missCount := votePenaltyCounter.MissCount

		// validate the total voting amount (success, abstain and miss)
		totalVotes := successCount + abstainCount + missCount
		if totalVotes == 0 {
			ctx.Logger().Error("zero votes in penalty counter, this should never happen")
			return false, nil
		}

		// rate = successVotes / total votes
		validVoteRate := math.LegacyNewDec(int64(successCount)).QuoInt64(int64(totalVotes))

		// penalize the validator whose the valid rate is smaller than the min threshold
		if validVoteRate.LT(minValidPerWindow) {
			validator, err := k.StakingKeeper.Validator(ctx, operator) // get validator
			if err != nil {
				panic(err)
			}
			if validator.IsBonded() && !validator.IsJailed() { // only bonded validators can be slashed
				consAddr, err := validator.GetConsAddr()
				if err != nil {
					panic(err)
				}

				consensusPower := validator.GetConsensusPower(powerReduction)
				_, err = k.StakingKeeper.Slash(ctx, consAddr, distributionHeight, consensusPower, slashFraction) // slash validator
				if err != nil {
					return true, err
				}
				err = k.StakingKeeper.Jail(ctx, consAddr) // Jail validator
				if err != nil {
					return true, err
				}
			}
		}

		// Emit an event with the validator address and its voting data
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(types.EventTypeEndSlashWindow,
				sdk.NewAttribute(types.AttributeKeyOperator, operator.String()),
				sdk.NewAttribute(types.AttributeKeyMissCount, strconv.FormatUint(missCount, 10)),
				sdk.NewAttribute(types.AttributeKeyAbstainCount, strconv.FormatUint(abstainCount, 10)),
				sdk.NewAttribute(types.AttributeKeySuccessCount, strconv.FormatUint(successCount, 10)),
			),
		)

		// Reset voting counter
		err := k.VotePenaltyCounter.Remove(ctx, operator)
		return false, err
	})
	return err
}
