package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/kiichain/kiichain/v1/x/rewards/types"
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the distribution MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (k msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if err := k.validateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	if err := msg.Params.ValidateBasic(); err != nil {
		return nil, err
	}

	if err := k.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (k msgServer) FundPool(ctx context.Context, msg *types.MsgFundPool) (*types.MsgFundPoolResponse, error) {
	depositor, err := k.accountKeeper.AddressCodec().StringToBytes(msg.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid depositor address: %s", err)
	}

	if err := validateAmount(msg.Amount); err != nil {
		return nil, err
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	if params.TokenDenom != msg.Amount.Denom {
		return nil, fmt.Errorf("Denom %s does not match expected denom: %s", msg.Amount.Denom, params.TokenDenom)
	}

	if err := k.Keeper.FundCommunityPool(ctx, msg.Amount, depositor); err != nil {
		return nil, err
	}

	return &types.MsgFundPoolResponse{}, nil
}

func (k msgServer) ExtendReward(ctx context.Context, msg *types.MsgExtendReward) (*types.MsgExtendRewardResponse, error) {
	if err := k.validateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	// Validate amt
	if err := validateAmount(msg.ExtraAmount); err != nil {
		return nil, err
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	if params.TokenDenom != msg.ExtraAmount.Denom {
		return nil, fmt.Errorf("Denom %s does not match expected denom: %s", msg.ExtraAmount.Denom, params.TokenDenom)
	}

	// Validate time
	// Should only time extentions be allowed? I.e do not allow reducing the time
	if err := validateTime(msg.EndTime); err != nil {
		return nil, err
	}

	// Check if funds exist (community pool funds - to be released > extra amount)
	if err := k.fundsAvailable(ctx, msg.ExtraAmount); err != nil {
		return nil, err
	}

	// Do actual work
	if err := k.Keeper.ExtendReward(ctx, msg.ExtraAmount, msg.EndTime); err != nil {
		return nil, err
	}

	return &types.MsgExtendRewardResponse{}, nil
}

// validateAuthority checks if address authority is valid and same as expected
func (k *Keeper) validateAuthority(authority string) error {
	if _, err := k.accountKeeper.AddressCodec().StringToBytes(authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	if k.authority != authority {
		return errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, authority)
	}

	return nil
}

// validateAmount check if amount is a valid coin
func validateAmount(amount sdk.Coin) error {
	if err := amount.Validate(); err != nil {
		return errors.Wrap(sdkerrors.ErrInvalidCoins, amount.String())
	}

	return nil
}

// validateTime checks if time is in the future
func validateTime(endTime time.Time) error {
	if endTime.After(time.Now()) {
		return fmt.Errorf("End time %s is not in the future", endTime)
	}

	return nil
}

// fundsAvailable checks if the asked funds are available in the pool
func (k Keeper) fundsAvailable(ctx context.Context, amount sdk.Coin) error {
	// Fetch releaser
	releaser, err := k.RewardReleaser.Get(ctx)
	if err != nil {
		return err
	}
	// Check if releaser is active (means some amt of the pool is promised)
	if releaser.Active {
		// Sum the promised amt to the asked funds
		amount = amount.Add(releaser.TotalAmount.Sub(releaser.ReleasedAmount))
	}

	// Get reward pool
	rewardPool, err := k.RewardPool.Get(ctx)
	if err != nil {
		return err
	}

	// Check if it has more funds than requested
	poolAmount := rewardPool.CommunityPool.AmountOf(amount.Denom)
	if math.LegacyDec(amount.Amount).GTE(poolAmount) {
		return fmt.Errorf("Reward pool (%s) has less funds than requested (%s)", poolAmount, amount)
	}

	return nil
}
