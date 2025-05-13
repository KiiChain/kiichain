package wasmbinding

import (
	"encoding/json"
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v1/wasmbinding/tokenfactory"
	bindingtypes "github.com/kiichain/kiichain/v1/wasmbinding/tokenfactory/types"
)

// QueryPlugin is the query plugin for all cosmwasm bindings
type QueryPlugin struct {
	tokenfactoryHandler tokenfactory.QueryPlugin
}

// NewQueryPlugin returns a reference to a new QueryPlugin
func NewQueryPlugin(th *tokenfactory.QueryPlugin) *QueryPlugin {
	return &QueryPlugin{
		tokenfactoryHandler: *th,
	}
}

// CustomQuerier dispatches custom CosmWasm bindings queries.
func CustomQuerier(qp *QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		// Unmarshal the requests as a query wrapper to be broken down per module
		var contractQuery bindingtypes.Query
		if err := json.Unmarshal(request, &contractQuery); err != nil {
			return nil, errorsmod.Wrap(err, "Error parsing request data")
		}

		// Match the query under the module
		switch {
		case contractQuery.FullDenom != nil:
			creator := contractQuery.FullDenom.CreatorAddr
			subdenom := contractQuery.FullDenom.Subdenom

			fullDenom, err := tokenfactory.GetFullDenom(creator, subdenom)
			if err != nil {
				return nil, errorsmod.Wrap(err, "kii full denom query")
			}

			res := bindingtypes.FullDenomResponse{
				Denom: fullDenom,
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, errorsmod.Wrap(err, "failed to marshal FullDenomResponse")
			}

			return bz, nil

		case contractQuery.Admin != nil:
			res, err := qp.tokenfactoryHandler.GetTokenfactoryDenomAdmin(ctx, contractQuery.Admin.Denom)
			if err != nil {
				return nil, err
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, fmt.Errorf("failed to JSON marshal AdminResponse: %w", err)
			}

			return bz, nil

		case contractQuery.Metadata != nil:
			res, err := qp.tokenfactoryHandler.GetTokenfactoryMetadata(ctx, contractQuery.Metadata.Denom)
			if err != nil {
				return nil, err
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, fmt.Errorf("failed to JSON marshal MetadataResponse: %w", err)
			}

			return bz, nil

		case contractQuery.DenomsByCreator != nil:
			res, err := qp.tokenfactoryHandler.GetTokenfactoryDenomsByCreator(ctx, contractQuery.DenomsByCreator.Creator)
			if err != nil {
				return nil, err
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, fmt.Errorf("failed to JSON marshal DenomsByCreatorResponse: %w", err)
			}

			return bz, nil

		case contractQuery.Params != nil:
			res, err := qp.tokenfactoryHandler.GetTokenfactoryParams(ctx)
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
}
