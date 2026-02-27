package keeper

import (
	"errors"
	"strconv"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kiichain/kiichain/v7/x/oracle/types"
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
			// If validator not found, skip slashing
			if errors.Is(err, stakingtypes.ErrNoValidatorFound) {
				return false, nil
			}
			// Check for other errors
			if err != nil || validator == nil {
				k.Logger(ctx).Error("failed to get validator for slashing", "operator", operator.String(), "error", err)
				return false, err
			}

			// Only slash if the validator is bonded
			if validator.IsBonded() && !validator.IsJailed() {
				// Get consensus address
				consAddr, err := validator.GetConsAddr()
				if err != nil {
					k.Logger(ctx).Error("failed to get consensus address for slashing", "operator", operator.String(), "error", err)
					return false, err
				}

				// Calculate consensus power
				consensusPower := validator.GetConsensusPower(powerReduction)

				// Slash the validator
				_, err = k.StakingKeeper.Slash(ctx, consAddr, distributionHeight, consensusPower, slashFraction)
				if err != nil {
					k.Logger(ctx).Error("failed to slash validator", "operator", operator.String(), "error", err)
					return true, err
				}

				// Jail validator
				err = k.StakingKeeper.Jail(ctx, consAddr)
				if err != nil {
					k.Logger(ctx).Error("failed to jail validator", "operator", operator.String(), "error", err)
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
