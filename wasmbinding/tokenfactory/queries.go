package tokenfactory

import (
	"context"
	"encoding/json"
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	tfbindingtypes "github.com/kiichain/kiichain/v2/wasmbinding/tokenfactory/types"
	"github.com/kiichain/kiichain/v2/wasmbinding/utils"
	tokenfactorykeeper "github.com/kiichain/kiichain/v2/x/tokenfactory/keeper"
)

// QueryPlugin is a custom query plugin for the wasm module for the token factory
type QueryPlugin struct {
	bankKeeper         bankkeeper.Keeper
	tokenFactoryKeeper *tokenfactorykeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin for the token factory module
func NewQueryPlugin(b bankkeeper.Keeper, tfk *tokenfactorykeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{
		bankKeeper:         b,
		tokenFactoryKeeper: tfk,
	}
}

// HandleTokenFactoryQuery is a custom querier for the token factory module
func (qp *QueryPlugin) HandleTokenFactoryQuery(ctx sdk.Context, tokenfactoryQuery tfbindingtypes.Query) ([]byte, error) {
	// Match the query under the module
	switch {
	// The query is a full denom query
	case tokenfactoryQuery.FullDenom != nil:
		creator := tokenfactoryQuery.FullDenom.CreatorAddr
		subdenom := tokenfactoryQuery.FullDenom.Subdenom

		fullDenom, err := GetFullDenom(creator, subdenom)
		if err != nil {
			return nil, errorsmod.Wrap(err, "kii full denom query")
		}

		res := tfbindingtypes.FullDenomResponse{
			Denom: fullDenom,
		}

		bz, err := json.Marshal(res)
		if err != nil {
			return nil, errorsmod.Wrap(err, "failed to marshal FullDenomResponse")
		}

		return bz, nil

		// The query is a denom admin query
	case tokenfactoryQuery.Admin != nil:
		res, err := qp.GetTokenfactoryDenomAdmin(ctx, tokenfactoryQuery.Admin.Denom)
		if err != nil {
			return nil, err
		}

		bz, err := json.Marshal(res)
		if err != nil {
			return nil, fmt.Errorf("failed to JSON marshal AdminResponse: %w", err)
		}

		return bz, nil

		// The query is a metadata query
	case tokenfactoryQuery.Metadata != nil:
		res, err := qp.GetTokenfactoryMetadata(ctx, tokenfactoryQuery.Metadata.Denom)
		if err != nil {
			return nil, err
		}

		bz, err := json.Marshal(res)
		if err != nil {
			return nil, fmt.Errorf("failed to JSON marshal MetadataResponse: %w", err)
		}

		return bz, nil

		// The query is a denoms by creator query
	case tokenfactoryQuery.DenomsByCreator != nil:
		res, err := qp.GetTokenfactoryDenomsByCreator(ctx, tokenfactoryQuery.DenomsByCreator.Creator)
		if err != nil {
			return nil, err
		}

		bz, err := json.Marshal(res)
		if err != nil {
			return nil, fmt.Errorf("failed to JSON marshal DenomsByCreatorResponse: %w", err)
		}

		return bz, nil

		// The query is a params query
	case tokenfactoryQuery.Params != nil:
		res, err := qp.GetTokenfactoryParams(ctx)
		if err != nil {
			return nil, err
		}

		bz, err := json.Marshal(res)
		if err != nil {
			return nil, fmt.Errorf("failed to JSON marshal ParamsResponse: %w", err)
		}

		return bz, nil

	default:
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown token factory query variant"}
	}
}

// GetTokenfactoryDenomAdmin is a query to get denom admin
func (qp QueryPlugin) GetTokenfactoryDenomAdmin(ctx context.Context, denom string) (*tfbindingtypes.AdminResponse, error) {
	metadata, err := qp.tokenFactoryKeeper.GetAuthorityMetadata(sdk.UnwrapSDKContext(ctx), denom)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin for denom: %s", denom)
	}
	return &tfbindingtypes.AdminResponse{Admin: metadata.Admin}, nil
}

// GetTokenfactoryDenomsByCreator is a query to get denoms by creator
func (qp QueryPlugin) GetTokenfactoryDenomsByCreator(ctx context.Context, creator string) (*tfbindingtypes.DenomsByCreatorResponse, error) {
	// TODO: validate creator address
	denoms := qp.tokenFactoryKeeper.GetDenomsFromCreator(sdk.UnwrapSDKContext(ctx), creator)
	return &tfbindingtypes.DenomsByCreatorResponse{Denoms: denoms}, nil
}

// GetTokenfactoryMetadata is a query to get metadata for a denom
func (qp QueryPlugin) GetTokenfactoryMetadata(ctx context.Context, denom string) (*tfbindingtypes.MetadataResponse, error) {
	metadata, found := qp.bankKeeper.GetDenomMetaData(ctx, denom)
	var parsed *tfbindingtypes.Metadata
	if found {
		parsed = SdkMetadataToWasm(metadata)
	}
	return &tfbindingtypes.MetadataResponse{Metadata: parsed}, nil
}

// GetTokenfactoryParams is a query to get token factory params
func (qp QueryPlugin) GetTokenfactoryParams(ctx context.Context) (*tfbindingtypes.ParamsResponse, error) {
	params := qp.tokenFactoryKeeper.GetParams(sdk.UnwrapSDKContext(ctx))
	return &tfbindingtypes.ParamsResponse{
		Params: tfbindingtypes.Params{
			DenomCreationFee: utils.ConvertSdkCoinsToWasmCoins(params.DenomCreationFee),
		},
	}, nil
}
