package keeper

import (
	"context"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
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

	// Set the params
	if err := ms.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	// Return the response
	return &types.MsgUpdateParamsResponse{}, nil
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
