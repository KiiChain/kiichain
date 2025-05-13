package tokenfactory

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	bindingtypes "github.com/kiichain/kiichain/v1/wasmbinding/types"
	"github.com/kiichain/kiichain/v1/wasmbinding/utils"
	tokenfactorykeeper "github.com/kiichain/kiichain/v1/x/tokenfactory/keeper"
)

// QueryPlugin is a custom query plugin for the wasm module for the token factory
type QueryPlugin struct {
	bankKeeper         bankkeeper.Keeper
	tokenFactoryKeeper *tokenfactorykeeper.Keeper
}

// QueryPlugin returns a reference to a new QueryPlugin for the token factory module
func NewQueryPlugin(b bankkeeper.Keeper, tfk *tokenfactorykeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{
		bankKeeper:         b,
		tokenFactoryKeeper: tfk,
	}
}

// GetTokenfactoryDenomAdmin is a query to get denom admin
func (qp QueryPlugin) GetTokenfactoryDenomAdmin(ctx context.Context, denom string) (*bindingtypes.AdminResponse, error) {
	metadata, err := qp.tokenFactoryKeeper.GetAuthorityMetadata(sdk.UnwrapSDKContext(ctx), denom)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin for denom: %s", denom)
	}
	return &bindingtypes.AdminResponse{Admin: metadata.Admin}, nil
}

// GetTokenfactoryDenomsByCreator is a query to get denoms by creator
func (qp QueryPlugin) GetTokenfactoryDenomsByCreator(ctx context.Context, creator string) (*bindingtypes.DenomsByCreatorResponse, error) {
	// TODO: validate creator address
	denoms := qp.tokenFactoryKeeper.GetDenomsFromCreator(sdk.UnwrapSDKContext(ctx), creator)
	return &bindingtypes.DenomsByCreatorResponse{Denoms: denoms}, nil
}

// GetTokenfactoryMetadata is a query to get metadata for a denom
func (qp QueryPlugin) GetTokenfactoryMetadata(ctx context.Context, denom string) (*bindingtypes.MetadataResponse, error) {
	metadata, found := qp.bankKeeper.GetDenomMetaData(ctx, denom)
	var parsed *bindingtypes.Metadata
	if found {
		parsed = SdkMetadataToWasm(metadata)
	}
	return &bindingtypes.MetadataResponse{Metadata: parsed}, nil
}

// GetTokenfactoryParams is a query to get token factory params
func (qp QueryPlugin) GetTokenfactoryParams(ctx context.Context) (*bindingtypes.ParamsResponse, error) {
	params := qp.tokenFactoryKeeper.GetParams(sdk.UnwrapSDKContext(ctx))
	return &bindingtypes.ParamsResponse{
		Params: bindingtypes.Params{
			DenomCreationFee: utils.ConvertSdkCoinsToWasmCoins(params.DenomCreationFee),
		},
	}, nil
}
