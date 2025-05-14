package wasmbinding

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v1/wasmbinding/tokenfactory"
	tfbindingtypes "github.com/kiichain/kiichain/v1/wasmbinding/tokenfactory/types"
)

// KiichainQuery is the query type for all cosmwasm bindings
type KiichainQuery struct {
	TokenFactory *tfbindingtypes.Query `json:"token_factory,omitempty"`
}

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
		var contractQuery KiichainQuery
		if err := json.Unmarshal(request, &contractQuery); err != nil {
			return nil, errorsmod.Wrap(err, "Error parsing request data")
		}

		// Match the query under the module
		switch {
		case contractQuery.TokenFactory != nil:
			// Call the token factory custom querier
			return qp.tokenfactoryHandler.HandleTokenFactoryQuery(ctx, *contractQuery.TokenFactory)
		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown query variant"}
		}
	}
}
