package keeper

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	erc20types "github.com/cosmos/evm/x/erc20/types"

	"github.com/kiichain/kiichain/v7/x/feeabstraction/types"
)

// MsgServer defines the keeper MsgServer wrapper
type MsgServer struct {
	Keeper
}

var _ types.MsgServer = MsgServer{}

// NewMsgServer returns the keeper message server
func NewMsgServer(k Keeper) types.MsgServer {
	return &MsgServer{Keeper: k}
}

// UpdateParams updates the module params though a proposal
func (ms MsgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	// Check the authority
	if err := ms.validateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	// Validate the message
	if msg == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("msg cannot be nil")
	}
	if err := msg.Validate(); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid message: %s", err)
	}

	// Validate the twap lookback window
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if err := ms.oracleKeeper.ValidateLookBackSeconds(sdkCtx, msg.Params.TwapLookbackWindow); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid twap lookback window: %s", err)
	}

	// Validate NativeOracleDenom is registered as an oracle vote target
	voteTargets, err := ms.oracleKeeper.GetVoteTargets(sdkCtx)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("failed to get oracle vote targets: %s", err)
	}
	found := false
	for _, denom := range voteTargets {
		if denom == msg.Params.NativeOracleDenom {
			found = true
			break
		}
	}
	if !found {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("native oracle denom %s is not registered as an oracle vote target", msg.Params.NativeOracleDenom)
	}

	// Validate NativeOracleDenom is in the oracle whitelist
	whitelist, err := ms.oracleKeeper.GetWhitelist(sdkCtx)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("failed to get oracle whitelist: %s", err)
	}
	if !whitelist.Contains(msg.Params.NativeOracleDenom) {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("native oracle denom %s is not in the oracle whitelist", msg.Params.NativeOracleDenom)
	}

	// Set the params
	if err := ms.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	// Return the response
	return &types.MsgUpdateParamsResponse{}, nil
}

// UpdateFeeTokens updates the fee tokens through a proposal
func (ms MsgServer) UpdateFeeTokens(ctx context.Context, msg *types.MsgUpdateFeeTokens) (*types.MsgUpdateFeeTokensResponse, error) {
	// Check the authority
	if err := ms.validateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	// Validate the message
	if msg == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("msg cannot be nil")
	}
	if err := msg.Validate(); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid message: %s", err)
	}

	// Check if all the oracle denoms are registered on the oracle module
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	voteTargets, err := ms.oracleKeeper.GetVoteTargets(sdkCtx)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("failed to get oracle vote targets: %s", err)
	}
	voteTargetMap := make(map[string]struct{}, len(voteTargets))
	for _, denom := range voteTargets {
		voteTargetMap[denom] = struct{}{}
	}
	for _, feeToken := range msg.Tokens.Items {
		if _, ok := voteTargetMap[feeToken.OracleDenom]; !ok {
			return nil, sdkerrors.ErrInvalidRequest.Wrapf("fee token denom %s is not registered on the oracle module", feeToken.OracleDenom)
		}
	}

	// Validate that the registered Decimals match the canonical decimals for each token
	for _, feeToken := range msg.Tokens.Items {
		if err := ms.validateFeeTokenDecimals(sdkCtx, feeToken); err != nil {
			return nil, sdkerrors.ErrInvalidRequest.Wrapf("fee token %s decimals mismatch: %s", feeToken.Denom, err)
		}
	}

	// Zero out all token prices so BeginBlocker adopts the current TWAP immediately
	// rather than being clamped from a stale governance-submitted price.
	// Clamp skip can be found here on /x/feeabstraction/types/fee.go#L55
	resetItems := make([]types.FeeTokenMetadata, len(msg.Tokens.Items))
	for i, token := range msg.Tokens.Items {
		resetItems[i] = types.NewFeeTokenMetadata(token.Denom, token.OracleDenom, token.Decimals, math.LegacyZeroDec())
	}
	resetFeeTokens := types.NewFeeTokenMetadataCollection(resetItems...)

	// Update the fee tokens
	if err := ms.FeeTokens.Set(ctx, *resetFeeTokens); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("failed to update fee tokens: %s", err)
	}

	// Return the response
	return &types.MsgUpdateFeeTokensResponse{}, nil
}

// validateFeeTokenDecimals checks that the Decimals field on a proposed fee token
// matches the ERC20 or bank record
func (ms MsgServer) validateFeeTokenDecimals(ctx sdk.Context, token types.UpdateTokenMetadata) error {
	// ERC20 path
	if strings.HasPrefix(token.Denom, erc20types.Erc20NativeCoinDenomPrefix) {
		hexAddr := strings.TrimPrefix(token.Denom, erc20types.Erc20NativeCoinDenomPrefix)

		// Look up the token pair to confirm it is registered.
		pairID := ms.erc20Keeper.GetTokenPairID(ctx, token.Denom)
		if len(pairID) == 0 {
			return sdkerrors.ErrUnknownAddress.Wrapf("no ERC20 pair registered for %s", token.Denom)
		}

		// Query decimals() from the contract.
		erc20Data, err := ms.erc20Keeper.QueryERC20(ctx, common.HexToAddress(hexAddr))
		if err != nil {
			return errors.Wrapf(err, "failed to query ERC20 decimals for %s", token.Denom)
		}

		if uint32(erc20Data.Decimals) != token.Decimals {
			return sdkerrors.ErrInvalidRequest.Wrapf(
				"declared decimals %d does not match contract decimals %d",
				token.Decimals, erc20Data.Decimals,
			)
		}
		return nil
	}

	// Bank-native path
	meta, found := ms.bankKeeper.GetDenomMetaData(ctx, token.Denom)
	if !found {
		// No metadata registered; skip the check.
		return nil
	}

	// Takes maximum exponent across all DenomUnits.
	var maxExp uint32
	for _, unit := range meta.DenomUnits {
		if unit.Exponent > maxExp {
			maxExp = unit.Exponent
		}
	}

	if maxExp != token.Decimals {
		return sdkerrors.ErrInvalidRequest.Wrapf(
			"declared decimals %d does not match bank metadata decimals %d",
			token.Decimals, maxExp,
		)
	}
	return nil
}

// validateAuthority checks if address authority is valid and same as expected
func (ms MsgServer) validateAuthority(authority string) error {
	// Parse the authority as a acc address
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	// Compare the authorities
	if ms.authority != authority {
		return errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, authority)
	}

	return nil
}
