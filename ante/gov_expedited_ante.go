package ante

import (
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	xerrors "github.com/kiichain/kiichain/v7/x/types/errors"
)

var expeditedPropDecoratorEnabled = true

// SetExpeditedProposalsEnabled toggles the expedited proposals decorator on/off.
// Should only be used in testing - this is to allow simtests to pass.
func SetExpeditedProposalsEnabled(val bool) {
	expeditedPropDecoratorEnabled = val
}

var expeditedPropsWhitelist = map[string]struct{}{
	"/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade": {},
	"/cosmos.upgrade.v1beta1.MsgCancelUpgrade":   {},
}

// GovExpeditedProposalsDecorator rejects expedited proposals whose messages are not whitelisted.
type GovExpeditedProposalsDecorator struct {
	cdc codec.BinaryCodec
}

// NewGovExpeditedProposalsDecorator returns a new GovExpeditedProposalsDecorator.
func NewGovExpeditedProposalsDecorator(cdc codec.BinaryCodec) GovExpeditedProposalsDecorator {
	return GovExpeditedProposalsDecorator{
		cdc: cdc,
	}
}

// AnteHandle checks if the proposal is whitelisted for expedited voting.
// Only proposals submitted using "kiichaind tx gov submit-proposal" can be expedited.
// Legacy proposals submitted using "kiichaind tx gov submit-legacy-proposal" cannot be marked as expedited.
func (g GovExpeditedProposalsDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if expeditedPropDecoratorEnabled {
		if err := g.validateMsgs(tx.GetMsgs()); err != nil {
			return ctx, err
		}
	}
	return next(ctx, tx, simulate)
}

// validateMsgs validates each message, unwrapping authz.MsgExec to prevent whitelist bypass.
func (g GovExpeditedProposalsDecorator) validateMsgs(msgs []sdk.Msg) error {
	for _, msg := range msgs {
		if execMsg, ok := msg.(*authz.MsgExec); ok {
			if err := g.validateAuthzExec(execMsg); err != nil {
				return err
			}
			continue
		}

		prop, ok := msg.(*govv1.MsgSubmitProposal)
		if !ok {
			continue
		}
		if prop.Expedited {
			if err := g.validateExpeditedGovProp(prop); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateAuthzExec unpacks the inner messages of an authz.MsgExec and validates them.
func (g GovExpeditedProposalsDecorator) validateAuthzExec(execMsg *authz.MsgExec) error {
	innerMsgs := make([]sdk.Msg, 0, len(execMsg.Msgs))
	for _, v := range execMsg.Msgs {
		var innerMsg sdk.Msg
		if err := g.cdc.UnpackAny(v, &innerMsg); err != nil {
			return errorsmod.Wrapf(xerrors.ErrInvalidExpeditedProposal, "cannot unmarshal authz exec msg (type %s): %v", v.TypeUrl, err)
		}
		innerMsgs = append(innerMsgs, innerMsg)
	}
	return g.validateMsgs(innerMsgs)
}

// isWhitelisted reports whether the given message type may be expedited.
func (g GovExpeditedProposalsDecorator) isWhitelisted(msgType string) bool {
	_, ok := expeditedPropsWhitelist[msgType]
	return ok
}

// validateExpeditedGovProp ensures every message in an expedited proposal is whitelisted.
func (g GovExpeditedProposalsDecorator) validateExpeditedGovProp(prop *govv1.MsgSubmitProposal) error {
	msgs := prop.GetMessages()
	if len(msgs) == 0 {
		return xerrors.ErrInvalidExpeditedProposal
	}
	for _, message := range msgs {
		// in case of legacy content submitted using govv1.MsgSubmitProposal
		if sdkMsg, isLegacy := message.GetCachedValue().(*govv1.MsgExecLegacyContent); isLegacy {
			if !g.isWhitelisted(sdkMsg.Content.TypeUrl) {
				return errorsmod.Wrapf(xerrors.ErrInvalidExpeditedProposal, "invalid Msg type: %s", sdkMsg.Content.TypeUrl)
			}
			continue
		}
		if !g.isWhitelisted(message.TypeUrl) {
			return errorsmod.Wrapf(xerrors.ErrInvalidExpeditedProposal, "invalid Msg type: %s", message.TypeUrl)
		}
	}
	return nil
}
