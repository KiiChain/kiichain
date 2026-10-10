package kiichain

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ baseapp.ProposalTxVerifier = &NoCheckProposalTxVerifier{}

// NoCheckProposalTxVerifier skips ante checks at proposal time. Krakatoa
// already validated every transaction in the proposal.
type NoCheckProposalTxVerifier struct {
	*baseapp.BaseApp
}

func NewNoCheckProposalTxVerifier(b *baseapp.BaseApp) *NoCheckProposalTxVerifier {
	return &NoCheckProposalTxVerifier{BaseApp: b}
}

func (txv *NoCheckProposalTxVerifier) PrepareProposalVerifyTx(tx sdk.Tx) ([]byte, error) {
	return txv.TxEncode(tx)
}
