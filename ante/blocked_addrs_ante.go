package ante

import (
	"strings"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	evmtypes "github.com/cosmos/evm/x/vm/types"

	xerrors "github.com/kiichain/kiichain/v7/x/types/errors"
)

// BlockedAddrDecorator rejects txs from or to addresses on the incident deny list.
// Enforcement of leftover funds is the bank send restriction; this helper is
// still used by Prepare/ProcessProposal to drop packed incident txs.
type BlockedAddrDecorator struct {
	cdc codec.BinaryCodec
}

// NewBlockedAddrDecorator returns a decorator that enforces the deny list.
func NewBlockedAddrDecorator(cdc codec.BinaryCodec) BlockedAddrDecorator {
	return BlockedAddrDecorator{cdc: cdc}
}

// AnteHandle rejects a tx that signs, sends, or calls a denied address.
func (d BlockedAddrDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if err := CheckBlockedTx(d.cdc, tx); err != nil {
		return ctx, err
	}
	return next(ctx, tx, simulate)
}

// CheckBlockedTx returns an error if tx uses a denied address as signer,
// bank sender/recipient, or MsgEthereumTx from/to.
func CheckBlockedTx(cdc codec.BinaryCodec, tx sdk.Tx) error {
	if sigTx, ok := tx.(authsigning.SigVerifiableTx); ok {
		signers, err := sigTx.GetSigners()
		if err == nil {
			for _, signer := range signers {
				if IsBlockedAccAddress(signer) {
					return blockedAddrErr(sdk.AccAddress(signer).String())
				}
			}
		}
	}
	return checkBlockedMsgs(cdc, tx.GetMsgs())
}

func checkBlockedMsgs(cdc codec.BinaryCodec, msgs []sdk.Msg) error {
	for _, msg := range msgs {
		if execMsg, ok := msg.(*authz.MsgExec); ok {
			if err := checkBlockedAuthzExec(cdc, execMsg); err != nil {
				return err
			}
			continue
		}
		if err := checkBlockedMsg(msg); err != nil {
			return err
		}
	}
	return nil
}

func checkBlockedAuthzExec(cdc codec.BinaryCodec, execMsg *authz.MsgExec) error {
	innerMsgs := make([]sdk.Msg, 0, len(execMsg.Msgs))
	for _, v := range execMsg.Msgs {
		var innerMsg sdk.Msg
		if err := cdc.UnpackAny(v, &innerMsg); err != nil {
			return errorsmod.Wrapf(xerrors.ErrUnauthorized, "cannot unmarshal authz exec msg (type %s): %v", v.TypeUrl, err)
		}
		innerMsgs = append(innerMsgs, innerMsg)
	}
	return checkBlockedMsgs(cdc, innerMsgs)
}

func checkBlockedMsg(msg sdk.Msg) error {
	switch m := msg.(type) {
	case *banktypes.MsgSend:
		if IsBlockedAddr(m.FromAddress) {
			return blockedAddrErr(m.FromAddress)
		}
		if IsBlockedAddr(m.ToAddress) {
			return blockedAddrErr(m.ToAddress)
		}
	case *banktypes.MsgMultiSend:
		for _, in := range m.Inputs {
			if IsBlockedAddr(in.Address) {
				return blockedAddrErr(in.Address)
			}
		}
		for _, out := range m.Outputs {
			if IsBlockedAddr(out.Address) {
				return blockedAddrErr(out.Address)
			}
		}
	case *evmtypes.MsgEthereumTx:
		if from := m.GetFrom(); IsBlockedAccAddress(from) {
			return blockedAddrErr(from.String())
		}
		if ethTx := m.AsTransaction(); ethTx != nil {
			if to := ethTx.To(); to != nil && IsBlockedAddr(to.Hex()) {
				return blockedAddrErr(to.Hex())
			}
		}
	}

	if hasSigners, ok := msg.(interface{ GetSigners() []sdk.AccAddress }); ok {
		for _, signer := range hasSigners.GetSigners() {
			if IsBlockedAccAddress(signer) {
				return blockedAddrErr(signer.String())
			}
		}
	}
	return nil
}

func blockedAddrErr(addr string) error {
	return errorsmod.Wrapf(errortypes.ErrUnauthorized, "address is blocked: %s", strings.ToLower(strings.TrimSpace(addr)))
}
