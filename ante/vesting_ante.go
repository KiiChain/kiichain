package ante

import (
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	xerrors "github.com/kiichain/kiichain/v7/x/types/errors"
)

// blockedVestingCreateMsgs is the set of vesting-module messages that open a
// new account with LockedCoins. Those accounts are the book the EVM locked-
// balance snapshot must stay consistent with; new creates are rejected until
// that path is safe.
var blockedVestingCreateMsgs = map[string]struct{}{
	sdk.MsgTypeURL(&sdkvesting.MsgCreateVestingAccount{}):         {},
	sdk.MsgTypeURL(&sdkvesting.MsgCreatePeriodicVestingAccount{}): {},
	sdk.MsgTypeURL(&sdkvesting.MsgCreatePermanentLockedAccount{}): {},
}

// VestingAccountCreationDecorator rejects messages that create vesting or
// permanently locked accounts, including when nested in authz.MsgExec.
type VestingAccountCreationDecorator struct {
	cdc codec.BinaryCodec
}

// NewVestingAccountCreationDecorator returns a decorator that blocks new
// vesting account creation.
func NewVestingAccountCreationDecorator(cdc codec.BinaryCodec) VestingAccountCreationDecorator {
	return VestingAccountCreationDecorator{cdc: cdc}
}

// AnteHandle rejects blocked vesting-create messages at any authz nesting depth.
func (d VestingAccountCreationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if err := d.validateMsgs(tx.GetMsgs()); err != nil {
		return ctx, err
	}
	return next(ctx, tx, simulate)
}

func (d VestingAccountCreationDecorator) validateMsgs(msgs []sdk.Msg) error {
	for _, msg := range msgs {
		if execMsg, ok := msg.(*authz.MsgExec); ok {
			if err := d.validateAuthzExec(execMsg); err != nil {
				return err
			}
			continue
		}

		typeURL := sdk.MsgTypeURL(msg)
		if _, blocked := blockedVestingCreateMsgs[typeURL]; blocked {
			return errorsmod.Wrapf(xerrors.ErrUnauthorized, "vesting account creation is disabled: %s", typeURL)
		}
	}
	return nil
}

func (d VestingAccountCreationDecorator) validateAuthzExec(execMsg *authz.MsgExec) error {
	innerMsgs := make([]sdk.Msg, 0, len(execMsg.Msgs))
	for _, v := range execMsg.Msgs {
		var innerMsg sdk.Msg
		if err := d.cdc.UnpackAny(v, &innerMsg); err != nil {
			return errorsmod.Wrapf(xerrors.ErrUnauthorized, "cannot unmarshal authz exec msg (type %s): %v", v.TypeUrl, err)
		}
		innerMsgs = append(innerMsgs, innerMsg)
	}
	return d.validateMsgs(innerMsgs)
}
