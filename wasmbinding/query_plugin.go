package wasmbinding

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v2/wasmbinding/bech32"
	bech32bindingtypes "github.com/kiichain/kiichain/v2/wasmbinding/bech32/types"
	"github.com/kiichain/kiichain/v2/wasmbinding/evm"
	evmbindingtypes "github.com/kiichain/kiichain/v2/wasmbinding/evm/types"
	"github.com/kiichain/kiichain/v2/wasmbinding/tokenfactory"
	tfbindingtypes "github.com/kiichain/kiichain/v2/wasmbinding/tokenfactory/types"
)

// KiichainQuery is the query type for all cosmwasm bindings
type KiichainQuery struct {
	TokenFactory *tfbindingtypes.Query     `json:"token_factory,omitempty"`
	EVM          *evmbindingtypes.Query    `json:"evm,omitempty"`
	Bech32       *bech32bindingtypes.Query `json:"bech32,omitempty"`
}

// QueryPlugin is the query plugin for all cosmwasm bindings
type QueryPlugin struct {
	tokenfactoryHandler tokenfactory.QueryPlugin
	evmHandler          evm.QueryPlugin
	bech32Handler       bech32.QueryPlugin
}

// NewQueryPlugin returns a reference to a new QueryPlugin
func NewQueryPlugin(th *tokenfactory.QueryPlugin, evm *evm.QueryPlugin, bech32 *bech32.QueryPlugin) *QueryPlugin {
	return &QueryPlugin{
		tokenfactoryHandler: *th,
		evmHandler:          *evm,
		bech32Handler:       *bech32,
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
		case contractQuery.EVM != nil:
			// Call the EVM custom querier
			return qp.evmHandler.HandleEVMQuery(ctx, *contractQuery.EVM)
		case contractQuery.Bech32 != nil:
			// Call the bech32 custom querier
			return qp.bech32Handler.HandleBech32Query(ctx, *contractQuery.Bech32)
		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown query variant"}
		}
	}
}
