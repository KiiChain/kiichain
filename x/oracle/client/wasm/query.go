package wasm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	oraclekeeper "github.com/kiichain/kiichain/v2/x/oracle/keeper"
	"github.com/kiichain/kiichain/v2/x/oracle/types"
)

// OracleWasmQueryHandler represents a wasm bridge to execute queries on the query_server
type OracleWasmQueryHandler struct {
	oracleKeeper oraclekeeper.Keeper
}

// NewOracleWasmQueryHandler creates a new instance of OracleWasmQueryHandler
func NewOracleWasmQueryHandler(keeper *oraclekeeper.Keeper) *OracleWasmQueryHandler {
	return &OracleWasmQueryHandler{
		oracleKeeper: *keeper,
	}
}

// GetExchangeRates executes the ExchangeRates query on the query_server
func (handler OracleWasmQueryHandler) GetExchangeRates(ctx sdk.Context) (*types.QueryExchangeRatesResponse, error) {
	querier := oraclekeeper.NewQueryServer(handler.oracleKeeper)
	return querier.ExchangeRates(ctx, &types.QueryExchangeRatesRequest{})
}

// GetOracleTwaps executes the Twaps query on the query_server
func (handler OracleWasmQueryHandler) GetOracleTwaps(ctx sdk.Context, req *types.QueryTwapsRequest) (*types.QueryTwapsResponse, error) {
	querier := oraclekeeper.NewQueryServer(handler.oracleKeeper)
	return querier.Twaps(ctx, req)
}

// GetActives executes the Actives query on the query_server
func (handler OracleWasmQueryHandler) GetActives(ctx sdk.Context, req *types.QueryActivesRequest) (*types.QueryActivesResponse, error) {
	querier := oraclekeeper.NewQueryServer(handler.oracleKeeper)
	return querier.Actives(ctx, &types.QueryActivesRequest{})
}

// GetPriceSnapshotHistory executes the PriceSnapshotHistory query on the query_server
func (handler OracleWasmQueryHandler) GetPriceSnapshotHistory(ctx sdk.Context, req *types.QueryPriceSnapshotHistoryRequest) (*types.QueryPriceSnapshotHistoryResponse, error) {
	querier := oraclekeeper.NewQueryServer(handler.oracleKeeper)
	return querier.PriceSnapshotHistory(ctx, &types.QueryPriceSnapshotHistoryRequest{})
}

// GetFeederDelegation executes the FeederDelegation query on the query_server
func (handler OracleWasmQueryHandler) GetFeederDelegation(ctx sdk.Context, req *types.QueryFeederDelegationRequest) (*types.QueryFeederDelegationResponse, error) {
	querier := oraclekeeper.NewQueryServer(handler.oracleKeeper)
	return querier.FeederDelegation(ctx, &types.QueryFeederDelegationRequest{ValidatorAddr: req.ValidatorAddr})
}

// GetVotePenaltyCounter executes the VotePenaltyCounter query on the query_server
func (handler OracleWasmQueryHandler) GetVotePenaltyCounter(ctx sdk.Context, req *types.QueryVotePenaltyCounterRequest) (*types.QueryVotePenaltyCounterResponse, error) {
	querier := oraclekeeper.NewQueryServer(handler.oracleKeeper)
	return querier.VotePenaltyCounter(ctx, &types.QueryVotePenaltyCounterRequest{ValidatorAddr: req.ValidatorAddr})
}
