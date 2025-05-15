package evm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/gogo/status"
	"google.golang.org/grpc/codes"

	sdk "github.com/cosmos/cosmos-sdk/types"

	emvconfig "github.com/cosmos/evm/server/config"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	evmbindingtypes "github.com/kiichain/kiichain/v1/wasmbinding/evm/types"
)

// QueryPlugin is a custom query plugin for the EVM module
type QueryPlugin struct {
	evmKeeper *evmkeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin for the EVM module
func NewQueryPlugin(ek *evmkeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{
		evmKeeper: ek,
	}
}

// HandleEVMQuery is a custom querier for the EVM module
func (qp *QueryPlugin) HandleEVMQuery(ctx sdk.Context, evmQuery evmbindingtypes.Query) ([]byte, error) {
	// Match the query under the module
	switch {
	case evmQuery.EthCall != nil:
		res, err := qp.HandleEthCall(ctx, evmQuery.EthCall)
		if err != nil && (err.Error() == vm.ErrExecutionReverted.Error()) {
			return nil, ErrExecutionReverted
		}
		if err != nil {
			return nil, ErrExecutingEthCall
		}

		bz, err := json.Marshal(res)
		if err != nil {
			return nil, fmt.Errorf("failed to JSON marshal Eth Call: %w", err)
		}

		return bz, nil
	default:
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown token factory query variant"}
	}
}

// HandleEthCall handles the EthCall query
func (qp *QueryPlugin) HandleEthCall(ctx sdk.Context, call *evmbindingtypes.EthCall) (*evmbindingtypes.EthCallResponse, error) {
	// Parse the to and the data
	to := common.HexToAddress(call.Contract)
	data, err := hexutil.Decode(call.Data)
	if err != nil {
		return nil, err
	}

	// Build the arguments
	args := evmtypes.TransactionArgs{
		To:   &to,
		Data: (*hexutil.Bytes)(&data),
	}

	// Parse the arguments
	bz, err := json.Marshal(&args)
	if err != nil {
		return nil, err
	}

	// Build the request
	req := evmtypes.EthCallRequest{
		Args:            bz,
		GasCap:          emvconfig.DefaultGasCap,
		ChainId:         int64(qp.evmKeeper.GetParams(ctx).ChainConfig.ChainId),
		ProposerAddress: sdk.ConsAddress(ctx.BlockHeader().ProposerAddress),
	}

	// Build a timeout and wrap the context
	timeout := emvconfig.DefaultEVMTimeout
	var cancel context.CancelFunc
	var ctxWrapped context.Context
	ctxWrapped, cancel = context.WithTimeout(ctx, timeout)

	// Make sure the context is canceled when the call has completed
	// this makes sure resources are cleaned up.
	defer cancel()

	// Call the EVM keeper
	res, err := qp.evmKeeper.EthCall(ctxWrapped, &req)
	if err != nil {
		return nil, err
	}

	// Handle reverts
	if err = handleRevertError(res.VmError, res.Ret); err != nil {
		return nil, err
	}

	// Get the response
	return &evmbindingtypes.EthCallResponse{Data: hexutil.Encode(res.Ret)}, nil
}

// handleRevertError returns revert related error
// This code is from the EVM module
func handleRevertError(vmError string, ret []byte) error {
	if len(vmError) > 0 {
		if vmError != vm.ErrExecutionReverted.Error() {
			return status.Error(codes.Internal, vmError)
		}
		if len(ret) == 0 {
			return errors.New(vmError)
		}
		return evmtypes.NewExecErrorWithReason(ret)
	}
	return nil
}
