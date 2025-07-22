package keeper

import (
	"cosmossdk.io/math"

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
// This must be implemented
// TODO: implement me and the module
type Keeper struct {
	erc20Keeper types.Erc20Keeper
	bankKeeper  types.BankKeeper

	// Mocked fee prices
	mockPrices []FeePrice
}

// NewKeeper creates a new instance of the Keeper
// TODO: implement me and the module
func NewKeeper(erc20Keeper types.Erc20Keeper, bankKeeper types.BankKeeper) Keeper {
	return Keeper{
		erc20Keeper: erc20Keeper,
		bankKeeper:  bankKeeper,
	}
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

// GetParams returns the module params
// TODO: implement me and the module
func (k Keeper) GetParams(ctx sdk.Context) Params {
	return Params{
		DefaultDenom: "akii",
	}
}
