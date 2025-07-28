package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
)

var (
	MockErc20Address = "0x816644F8bc4633D268842628EB10ffC0AdcB6099"
	// The mock ERC20 denom
	MockErc20Denom = "erc20/" + MockErc20Address
	// The mock ERC20 price
	MockErc20Price = math.LegacyNewDecFromInt(math.NewInt(10)) // 10 uatom = 1 kii
)

// FeePrice is a generate
// TODO: implement me and the module
type FeePrice struct {
	Denom string
	Price math.LegacyDec
}

// Params define the module parameters
// TODO: implement me and the module
type Params struct {
	DefaultDenom string
}

// This is a mocked keeper for the fee abstraction module
type Keeper struct {
	// The chain codecs
	cdc codec.BinaryCodec

	// Modules used on the keeper
	bankKeeper  types.BankKeeper
	erc20Keeper types.Erc20Keeper

	// The governance authority
	authority string

	// The schema and the different entries on collections
	Schema collections.Schema
	Params collections.Item[types.Params]

	// Mocked fee prices
	mockPrices []FeePrice
}

// NewKeeper creates a new instance of the Keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	erc20Keeper types.Erc20Keeper, bankKeeper types.BankKeeper,
	authority string,
) Keeper {
	// Start a new schema builder
	sb := collections.NewSchemaBuilder(storeService)

	// Initialize the keeper
	k := Keeper{
		cdc:         cdc,
		erc20Keeper: erc20Keeper,
		bankKeeper:  bankKeeper,
		authority:   authority,
		Params:      collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
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

// SetFeePrices sets the fee prices
// This is a mocked function for the fee abstraction module
// TODO: implement me and the module
func (k *Keeper) SetFeePrices(ctx sdk.Context, prices []FeePrice) {
	k.mockPrices = prices
}

// GetFeeAbstractionTokens returns the fee abstraction tokens
// For now we use a mocked value
// TODO: implement me and the module
func (k Keeper) GetFeePrices(ctx sdk.Context) ([]FeePrice, error) {
	// Check if we have mocked values
	if k.mockPrices != nil {
		return k.mockPrices, nil
	}

	return []FeePrice{
		{
			Denom: MockErc20Denom,
			Price: MockErc20Price,
		},
	}, nil
}
