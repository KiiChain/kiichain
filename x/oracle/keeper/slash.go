package keeper

import (
	"errors"
	"strconv"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kiichain/kiichain/v7/x/oracle/types"
)

// SlashAndResetCounters calculate if the validator must be slashed if success votes / total votes
// is lower than MinValidPerWindow param. Then reset the vote penalty info
func (k Keeper) SlashAndResetCounters(ctx sdk.Context) error {
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

				// Redirect the slash amount to the community pool instead of burning
				_, err = k.slashToPool(ctx, consAddr, consensusPower, slashFraction)
				if err != nil {
					k.Logger(ctx).Error("failed to slash validator", "operator", operator.String(), "error", err)
					return true, err
				}

				// Jail validator if slashing is active
				if slashFraction.IsPositive() {
					err = k.StakingKeeper.Jail(ctx, consAddr)
					if err != nil {
						k.Logger(ctx).Error("failed to jail validator", "operator", operator.String(), "error", err)
						return true, err
					}
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

// slashToPool removes tokens from a validator and sends them to the community pool
// instead of burning, preserving total supply.
func (k Keeper) slashToPool(ctx sdk.Context, consAddr sdk.ConsAddress, consensusPower int64, slashFactor math.LegacyDec) (math.Int, error) {
	if slashFactor.IsZero() {
		return math.ZeroInt(), nil
	}

	// Calculate slash amount: same math as the staking module
	amount := k.StakingKeeper.TokensFromConsensusPower(ctx, consensusPower)
	slashAmount := math.LegacyNewDecFromInt(amount).Mul(slashFactor).TruncateInt()

	// Get the concrete validator (needed for RemoveValidatorTokens)
	validator, err := k.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	if err != nil {
		return math.ZeroInt(), err
	}

	// Clamp slash amount to the validator's available tokens
	tokensToBurn := math.MinInt(slashAmount, validator.Tokens)
	if tokensToBurn.IsZero() {
		return math.ZeroInt(), nil
	}

	// Remove tokens from the validator's bookkeeping without burning
	_, err = k.StakingKeeper.RemoveValidatorTokens(ctx, validator, tokensToBurn)
	if err != nil {
		return math.ZeroInt(), err
	}

	// Create the coin (value but with akii)
	bondDenom, err := k.StakingKeeper.BondDenom(ctx)
	if err != nil {
		return math.ZeroInt(), err
	}
	coins := sdk.NewCoins(sdk.NewCoin(bondDenom, tokensToBurn))
	// Send the slashed tokens from the bonded pool to the community pool
	bondedPoolAddr := authtypes.NewModuleAddress(stakingtypes.BondedPoolName)
	if err := k.distrKeeper.FundCommunityPool(ctx, coins, bondedPoolAddr); err != nil {
		return math.ZeroInt(), err
	}

	k.Logger(ctx).Info(
		"oracle slash redirected to community pool",
		"validator", consAddr.String(),
		"slash_factor", slashFactor.String(),
		"amount", tokensToBurn.String(),
	)

	return tokensToBurn, nil
}
