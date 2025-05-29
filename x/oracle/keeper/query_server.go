package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v1/x/oracle/types"
)

// queryServer struct that handlers the rpc request
type queryServer struct {
	Keeper Keeper
}

// Ensure the struct queryServer implement the QueryServer interface
var _ types.QueryServer = queryServer{}

// NewQueryServer returns a new instance of the QueryServer
func NewQueryServer(keepr Keeper) types.QueryServer {
	return queryServer{
		Keeper: keepr,
	}
}

// Params returns the oracle's params
func (qs queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	// Get the module's params from the keeper
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params, err := qs.Keeper.Params.Get(sdkCtx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: &params}, nil
}

// ExchangeRate returns the exchange rate specific by denom
func (qs queryServer) ExchangeRate(ctx context.Context, req *types.QueryExchangeRateRequest) (*types.QueryExchangeRateResponse, error) {
	// Validate request
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if len(req.Denom) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty denom")
	}

	// Get exchange rate by denom
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	exchangeRate, err := qs.Keeper.GetBaseExchangeRate(sdkCtx, req.Denom)
	if err != nil {
		return nil, err
	}

	// Prepare response
	response := &types.QueryExchangeRateResponse{
		OracleExchangeRate: &exchangeRate,
	}

	return response, nil
}

// ExchangeRates returns all exchange rates
func (qs queryServer) ExchangeRates(ctx context.Context, req *types.QueryExchangeRatesRequest) (*types.QueryExchangeRatesResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	exchangeRates := []types.DenomOracleExchangeRate{}
	qs.Keeper.IterateBaseExchangeRates(sdkCtx, func(denom string, exchangeRate types.OracleExchangeRate) (bool, error) {
		exchangeRates = append(exchangeRates, types.DenomOracleExchangeRate{Denom: denom, OracleExchangeRate: &exchangeRate})
		return false, nil
	})

	return &types.QueryExchangeRatesResponse{DenomOracleExchangeRate: exchangeRates}, nil
}

// Actives queries all denoms for which exchange rates exist
func (qs queryServer) Actives(ctx context.Context, req *types.QueryActivesRequest) (*types.QueryActivesResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	denomsActive := []string{}
	qs.Keeper.IterateBaseExchangeRates(sdkCtx, func(denom string, exchangeRate types.OracleExchangeRate) (bool, error) {
		denomsActive = append(denomsActive, denom)
		return false, nil
	})

	return &types.QueryActivesResponse{Actives: denomsActive}, nil
}

// VoteTargets queries the voting target list on current vote period
func (qs queryServer) VoteTargets(ctx context.Context, req *types.QueryVoteTargetsRequest) (*types.QueryVoteTargetsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &types.QueryVoteTargetsResponse{VoteTargets: qs.Keeper.GetVoteTargets(sdkCtx)}, nil
}

// PriceSnapshotHistory queries all snapshots
func (qs queryServer) PriceSnapshotHistory(ctx context.Context, req *types.QueryPriceSnapshotHistoryRequest) (*types.QueryPriceSnapshotHistoryResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get the snapshots available on the KVStore
	priceSnapshots := []types.PriceSnapshot{}
	qs.Keeper.IteratePriceSnapshots(sdkCtx, func(_ int64, snapshot types.PriceSnapshot) (bool, error) {
		priceSnapshots = append(priceSnapshots, snapshot)
		return false, nil
	})

	return &types.QueryPriceSnapshotHistoryResponse{PriceSnapshot: priceSnapshots}, nil
}

// Twaps queries the Time-weighted average price (TWAPs) whitin an specific period of time
func (qs queryServer) Twaps(ctx context.Context, req *types.QueryTwapsRequest) (*types.QueryTwapsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	twaps, err := qs.Keeper.CalculateTwaps(sdkCtx, req.LookbackSeconds)
	if err != nil {
		return nil, err
	}

	return &types.QueryTwapsResponse{OracleTwap: twaps}, err
}

// FeederDelegation queries the account data address assigned as a delegator by a validator
func (qs queryServer) FeederDelegation(ctx context.Context, req *types.QueryFeederDelegationRequest) (*types.QueryFeederDelegationResponse, error) {
	// Validate request information
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get the delegator by the Validator address
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	feederAddr := qs.Keeper.GetFeederDelegation(sdkCtx, valAddr).String()

	return &types.QueryFeederDelegationResponse{FeedAddr: feederAddr}, nil
}

// VotePenaltyCounter queries the validator penalty's counter information
func (qs queryServer) VotePenaltyCounter(ctx context.Context, req *types.QueryVotePenaltyCounterRequest) (*types.QueryVotePenaltyCounterResponse, error) {
	// Validate request information
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get the penalty counters by the validator address
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	missCount := qs.Keeper.GetMissCount(sdkCtx, valAddr)
	abstainCount := qs.Keeper.GetAbstainCount(sdkCtx, valAddr)
	successCount := qs.Keeper.GetSuccessCount(sdkCtx, valAddr)

	// Prepare response
	votePenaltyCounter := &types.VotePenaltyCounter{
		MissCount:    missCount,
		AbstainCount: abstainCount,
		SuccessCount: successCount,
	}

	return &types.QueryVotePenaltyCounterResponse{VotePenaltyCounter: votePenaltyCounter}, nil
}

// SlashWindow queries the slash window progress
func (qs queryServer) SlashWindow(ctx context.Context, req *types.QuerySlashWindowRequest) (*types.QuerySlashWindowResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params, err := qs.Keeper.Params.Get(sdkCtx)
	if err != nil {
		return nil, err
	}

	// The window progress is the number of vote periods that have been completed in the current slashing window.
	// With a vote period of 1, this will be equivalent to the number of blocks that have progressed in the slash window.
	windowProgress := (uint64(sdkCtx.BlockHeight()) % params.SlashWindow) / params.VotePeriod

	return &types.QuerySlashWindowResponse{WindowProgress: windowProgress}, nil
}
