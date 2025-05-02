package wasmd

import (
	wasmdkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/x/vm/core/vm"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

const (
	// QueryRawMethod is the method name for raw queries
	QueryRawMethod = "queryRaw"
	// QuerySmartMethod is the method name for smart queries
	QuerySmartMethod = "querySmart"
)

// QueryRaw is a precompile method that handles raw queries from the contract to the wasmd module
func (p *Precompile) QueryRaw(ctx sdk.Context, method *abi.Method, _ *vm.Contract, args []interface{}) ([]byte, error) {
	// Build the request from the arguments
	req, err := ParseQueryRawArgs(args)
	if err != nil {
		return nil, err
	}

	// Get the querier interface
	querier := wasmdkeeper.Querier(&p.wasmdKeeper)

	// Make the request
	res, err := querier.RawContractState(ctx, req)
	if err != nil {
		return nil, err
	}

	// Returnt the response
	return method.Outputs.Pack(res.Data)
}

// QuerySmart is a precompile method that handles smart queries from the contract to the wasmd module
func (p *Precompile) QuerySmart(ctx sdk.Context, method *abi.Method, _ *vm.Contract, args []interface{}) ([]byte, error) {
	// Build the request from the arguments
	req, err := ParseQuerySmartArgs(args)
	if err != nil {
		return nil, err
	}

	// Get the querier interface
	querier := wasmdkeeper.Querier(&p.wasmdKeeper)

	// Make the request
	res, err := querier.SmartContractState(ctx, req)
	if err != nil {
		return nil, err
	}

	// Returnt the response
	return method.Outputs.Pack(res.Data)
}
