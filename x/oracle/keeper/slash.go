package keeper

import (
	"strconv"

	cosmostelemetry "github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
)

// SlashAndResetCounters calculate if the validator must be slashed if success votes / total votes
// is lower than MinValidPerWindow param. Then reset the vote penalty info
func (k Keeper) SlashAndResetCounters(ctx sdk.Context) {
	height := ctx.BlockHeight()
	distributionHeight := height - sdk.ValidatorUpdateDelay - 1

	minValidPerWindow := k.MinValidPerWindow(ctx) // get from params
	slashFraction := k.SlashFraction(ctx)         // get from params
	powerReduction := k.StakingKeeper.PowerReduction(ctx)

	// Iterate each voting result per validator
	k.IterateVotePenaltyCounters(ctx, func(operator sdk.ValAddress, votePenaltyCounter types.VotePenaltyCounter) bool {
		successCount := votePenaltyCounter.SuccessCount
		abstainCount := votePenaltyCounter.AbstainCount
		missCount := votePenaltyCounter.MissCount

		// validate the total voting amount (success, abstain and miss)
		totalVotes := successCount + abstainCount + missCount
		if totalVotes == 0 {
			ctx.Logger().Error("zero votes in penalty counter, this should never happen")
			return false
		}

		// rate = successVotes / total votes
		validVoteRate := sdk.NewDec(int64(successCount)).QuoInt64(int64(totalVotes))

		// penalize the validator whose the valid rate is smaller than the min threshold
		if validVoteRate.LT(minValidPerWindow) {
			validator := k.StakingKeeper.Validator(ctx, operator) // get validator
			if validator.IsBonded() && !validator.IsJailed() {    // only bonded validators can be slashed
				consAddr, err := validator.GetConsAddr()
				if err != nil {
					panic(err)
				}

				consensusPower := validator.GetConsensusPower(powerReduction)
				k.StakingKeeper.Slash(ctx, consAddr, distributionHeight, consensusPower, slashFraction) // slash validator
				k.StakingKeeper.Jail(ctx, consAddr)                                                     // Jail validator
				cosmostelemetry.IncrValidatorSlashedCounter(consAddr.String(), "oracle")
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
		k.DeleteVotePenaltyCounter(ctx, operator)
		return false
	})
}
