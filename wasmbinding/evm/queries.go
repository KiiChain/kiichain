package evm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/gogo/status"
	"google.golang.org/grpc/codes"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/evm/contracts"
	emvconfig "github.com/cosmos/evm/server/config"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	evmbindingtypes "github.com/kiichain/kiichain/v2/wasmbinding/evm/types"
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
	case evmQuery.ERC20Information != nil:
		res, err := qp.HandleERC20Information(ctx, evmQuery.ERC20Information)
		if err != nil {
			return nil, fmt.Errorf("failed to handle ERC20 information: %w", err)
		}

		bz, err := json.Marshal(res)
		if err != nil {
			return nil, fmt.Errorf("failed to JSON marshal ERC20 information: %w", err)
		}
		return bz, nil
	case evmQuery.ERC20Balance != nil:
		res, err := qp.HandleERC20Balance(ctx, evmQuery.ERC20Balance)
		if err != nil {
			return nil, fmt.Errorf("failed to handle ERC20 balance: %w", err)
		}
		bz, err := json.Marshal(res)
		if err != nil {
			return nil, fmt.Errorf("failed to JSON marshal ERC20 balance: %w", err)
		}
		return bz, nil
	case evmQuery.ERC20Allowance != nil:
		res, err := qp.HandleERC20Allowance(ctx, evmQuery.ERC20Allowance)
		if err != nil {
			return nil, fmt.Errorf("failed to handle ERC20 allowance: %w", err)
		}
		bz, err := json.Marshal(res)
		if err != nil {
			return nil, fmt.Errorf("failed to JSON marshal ERC20 allowance: %w", err)
		}
		return bz, nil
	default:
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown token factory query variant"}
	}
}

// HandleEthCall handles the EthCall query
func (qp *QueryPlugin) HandleEthCall(ctx sdk.Context, call *evmbindingtypes.EthCallRequest) (*evmbindingtypes.EthCallResponse, error) {
	// Prepare the request data
	chainID := qp.evmKeeper.GetParams(ctx).ChainConfig.ChainId
	proposer := ctx.BlockHeader().ProposerAddress
	to := common.HexToAddress(call.Contract)

	// Parse the to and the data
	data, err := hexutil.Decode(call.Data)
	if err != nil {
		return nil, err
	}

	// Build the eth call request
	req, err := buildEthCallRequest(to, data, chainID, proposer)
	if err != nil {
		return nil, err
	}

	// Build a timeout and wrap the context
	timeout := emvconfig.DefaultEVMTimeout
	ctxWrapped, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Call the EVM keeper
	res, err := qp.callEVMAndHandleRevertError(ctxWrapped, req)
	if err != nil {
		return nil, err
	}

	// Get the response
	return &evmbindingtypes.EthCallResponse{Data: hexutil.Encode(res.Ret)}, nil
}

// HandleERC20Information handles the ERC20Information query
// Since we are using a ABI, we don't need to timeout the call
func (qp *QueryPlugin) HandleERC20Information(ctx sdk.Context, call *evmbindingtypes.ERC20InformationRequest) (*evmbindingtypes.ERC20InformationResponse, error) {
	// Prepare the request data
	to := common.HexToAddress(call.Contract)
	erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI

	// Response object
	res := &evmbindingtypes.ERC20InformationResponse{}

	// Query the decimals
	callRes, err := qp.evmKeeper.CallEVM(ctx, erc20ABI, erc20types.ModuleAddress, to, false, "decimals")
	if err != nil {
		return nil, err
	}
	// Unpack the decimals
	unpacked, err := erc20ABI.Unpack("decimals", callRes.Ret)
	if err != nil {
		return nil, err
	}
	// Get the decimals
	decimals, ok := unpacked[0].(uint8)
	if !ok {
		return nil, errors.New("decimals is not a big.Int")
	}
	res.Decimals = decimals

	// Query the name
	callRes, err = qp.evmKeeper.CallEVM(ctx, erc20ABI, erc20types.ModuleAddress, to, false, "name")
	if err != nil {
		return nil, err
	}
	// Unpack the name
	unpacked, err = erc20ABI.Unpack("name", callRes.Ret)
	if err != nil {
		return nil, err
	}
	// Get the name
	name, ok := unpacked[0].(string)
	if !ok {
		return nil, errors.New("name is not a string")
	}
	res.Name = name

	// Query the symbol
	callRes, err = qp.evmKeeper.CallEVM(ctx, erc20ABI, erc20types.ModuleAddress, to, false, "symbol")
	if err != nil {
		return nil, err
	}
	// Unpack the symbol
	unpacked, err = erc20ABI.Unpack("symbol", callRes.Ret)
	if err != nil {
		return nil, err
	}
	// Get the symbol
	symbol, ok := unpacked[0].(string)
	if !ok {
		return nil, errors.New("symbol is not a string")
	}
	res.Symbol = symbol

	// Query the total supply
	callRes, err = qp.evmKeeper.CallEVM(ctx, erc20ABI, erc20types.ModuleAddress, to, false, "totalSupply")
	if err != nil {
		return nil, err
	}
	// Unpack the total supply
	unpacked, err = erc20ABI.Unpack("totalSupply", callRes.Ret)
	if err != nil {
		return nil, err
	}
	// Get the total supply
	totalSupply, ok := unpacked[0].(*big.Int)
	if !ok {
		return nil, errors.New("totalSupply is not a big.Int")
	}
	res.TotalSupply = totalSupply.String()

	return res, nil
}

// HandleERC20Balance handles the ERC20Balance query
func (qp *QueryPlugin) HandleERC20Balance(ctx sdk.Context, call *evmbindingtypes.ERC20BalanceRequest) (*evmbindingtypes.ERC20BalanceResponse, error) {
	// Prepare the request data
	to := common.HexToAddress(call.Contract)
	address := common.HexToAddress(call.Address)
	erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI

	// Response object
	res := &evmbindingtypes.ERC20BalanceResponse{}

	// Query the balance
	callRes, err := qp.evmKeeper.CallEVM(ctx, erc20ABI, erc20types.ModuleAddress, to, false, "balanceOf", address)
	if err != nil {
		return nil, err
	}
	// Unpack the balance
	unpacked, err := erc20ABI.Unpack("balanceOf", callRes.Ret)
	if err != nil {
		return nil, err
	}
	// Get the balance
	balance, ok := unpacked[0].(*big.Int)
	if !ok {
		return nil, errors.New("balance is not a big.Int")
	}
	res.Balance = balance.String()

	return res, nil
}

// HandleERC20Allowance handles the ERC20Allowance query
func (qp *QueryPlugin) HandleERC20Allowance(ctx sdk.Context, call *evmbindingtypes.ERC20AllowanceRequest) (*evmbindingtypes.ERC20AllowanceResponse, error) {
	// Prepare the request data
	to := common.HexToAddress(call.Contract)
	owner := common.HexToAddress(call.Owner)
	spender := common.HexToAddress(call.Spender)
	erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI

	// Response object
	res := &evmbindingtypes.ERC20AllowanceResponse{}

	// Query the allowance
	callRes, err := qp.evmKeeper.CallEVM(ctx, erc20ABI, erc20types.ModuleAddress, to, false, "allowance", owner, spender)
	if err != nil {
		return nil, err
	}
	// Unpack the allowance
	unpacked, err := erc20ABI.Unpack("allowance", callRes.Ret)
	if err != nil {
		return nil, err
	}
	// Get the allowance
	allowance, ok := unpacked[0].(*big.Int)
	if !ok {
		return nil, errors.New("allowance is not a big.Int")
	}
	res.Allowance = allowance.String()

	return res, nil
}

// callEVMAndHandleRevertError calls the EVM and handles revert errors
func (qp *QueryPlugin) callEVMAndHandleRevertError(ctx context.Context, req *evmtypes.EthCallRequest) (*evmtypes.MsgEthereumTxResponse, error) {
	// Call the EVM keeper
	res, err := qp.evmKeeper.EthCall(ctx, req)
	if err != nil {
		return nil, err
	}

	// Handle reverts
	if err = handleRevertError(res.VmError, res.Ret); err != nil {
		return nil, err
	}

	return res, nil
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

// buildEthCallRequest builds the EVM query call
func buildEthCallRequest(to common.Address, data hexutil.Bytes, chainID uint64, proposer []byte) (*evmtypes.EthCallRequest, error) {
	// Build the arguments
	args := evmtypes.TransactionArgs{
		To:   &to,
		Data: &data,
	}

	// Parse the arguments
	bz, err := json.Marshal(&args)
	if err != nil {
		return nil, err
	}

	// Return the request
	return &evmtypes.EthCallRequest{
		Args:            bz,
		GasCap:          emvconfig.DefaultGasCap,
		ChainId:         int64(chainID),
		ProposerAddress: sdk.ConsAddress(proposer),
	}, nil
}
