package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
)

// This is a mocked keeper for the fee abstraction module
type Keeper struct {
	// The chain codecs
	cdc codec.BinaryCodec

	// Modules used on the keeper
	bankKeeper   types.BankKeeper
	erc20Keeper  types.Erc20Keeper
	oracleKeeper types.OracleKeeper

	// The governance authority
	authority string

	// The schema and the different entries on collections
	Schema    collections.Schema
	Params    collections.Item[types.Params]
	FeeTokens collections.Item[types.FeeTokenMetadataCollection]
}

// NewKeeper creates a new instance of the Keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	erc20Keeper types.Erc20Keeper, bankKeeper types.BankKeeper, oracleKeeper types.OracleKeeper,
	authority string,
) Keeper {
	// Start a new schema builder
	sb := collections.NewSchemaBuilder(storeService)

	// Initialize the keeper
	k := Keeper{
		cdc:          cdc,
		erc20Keeper:  erc20Keeper,
		bankKeeper:   bankKeeper,
		oracleKeeper: oracleKeeper,
		authority:    authority,
		Params:       collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		FeeTokens:    collections.NewItem(sb, types.FeeTokensKey, "fee_tokens", codec.CollValue[types.FeeTokenMetadataCollection](cdc)),
	}

	// Build the schema
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	// Return the keeper
	return k
}

// GetAuthority returns the module authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns the logger for the module
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
