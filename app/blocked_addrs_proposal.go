package kiichain

import (
	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	kiiante "github.com/kiichain/kiichain/v7/ante"
)

// WrapPrepareProposal strips denied-address txs before the inner selector runs.
func WrapPrepareProposal(cdc codec.BinaryCodec, decoder sdk.TxDecoder, next sdk.PrepareProposalHandler) sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		req.Txs = filterBlockedProposalTxs(cdc, decoder, req.Txs)
		if next == nil {
			return &abci.ResponsePrepareProposal{Txs: req.Txs}, nil
		}
		return next(ctx, req)
	}
}

// WrapProcessProposal rejects a new proposal that still contains a denied-address tx.
// It does not return an error (that would stall a decided block if used in PreBlock).
func WrapProcessProposal(cdc codec.BinaryCodec, decoder sdk.TxDecoder, next sdk.ProcessProposalHandler) sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		if proposalContainsBlockedTx(cdc, decoder, req.Txs) {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}
		if next == nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
		}
		return next(ctx, req)
	}
}

func filterBlockedProposalTxs(cdc codec.BinaryCodec, decoder sdk.TxDecoder, txs [][]byte) [][]byte {
	out := make([][]byte, 0, len(txs))
	for _, raw := range txs {
		tx, err := decoder(raw)
		if err != nil {
			out = append(out, raw)
			continue
		}
		if err := kiiante.CheckBlockedTx(cdc, tx); err != nil {
			continue
		}
		out = append(out, raw)
	}
	return out
}

func proposalContainsBlockedTx(cdc codec.BinaryCodec, decoder sdk.TxDecoder, txs [][]byte) bool {
	for _, raw := range txs {
		tx, err := decoder(raw)
		if err != nil {
			continue
		}
		if err := kiiante.CheckBlockedTx(cdc, tx); err != nil {
			return true
		}
	}
	return false
}

func logBlockedFinalizeTxs(ctx sdk.Context, cdc codec.BinaryCodec, decoder sdk.TxDecoder, txs [][]byte) {
	for i, raw := range txs {
		tx, err := decoder(raw)
		if err != nil {
			continue
		}
		if err := kiiante.CheckBlockedTx(cdc, tx); err != nil {
			ctx.Logger().Error("blocked address tx in finalize block; ante will reject it", "index", i, "err", err)
		}
	}
}
