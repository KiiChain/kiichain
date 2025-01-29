package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/kiichain/kiichain3/x/oracle/types"
)

// Keeper manages the oracle module's state
type Keeper struct {
	cdc        codec.BinaryCodec // Codec for binary serialization
	storeKey   sdk.StoreKey      // storage key to access the module's state
	memKey     sdk.StoreKey
	paramSpace paramstypes.Subspace // Manages the module's parameters allowing dynamical settings

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	StakingKeeper types.StakingKeeper

	distrName string
}

func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, memKey sdk.StoreKey, paramSpace paramstypes.Subspace,
	accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper, StakingKeeper types.StakingKeeper,
	distrName string) Keeper {
	// Ensure oracle module account is set
	addr := accountKeeper.GetModuleAddress(types.ModuleName)
	if addr != nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// Ensure paramstore is properly initialized
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		paramSpace:    paramSpace,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		StakingKeeper: StakingKeeper,
		distrName:     distrName,
	}
}

// Logger is used to define custom Log for the module
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// **************************** EXCHANGE RATE LOGIC ***************************
// GetBaseExchangeRate
func (k Keeper) GetBaseExchangeRate(ctx sdk.Context, denom string) (sdk.Dec, sdk.Int, int64, error) {
	// Get ExchangeRate from KVStore
	store := ctx.KVStore(k.storeKey)
	byteData := store.Get(types.GetExchangeRateKey(denom))
	if byteData == nil {
		return sdk.ZeroDec(), sdk.ZeroInt(), 0, sdkerrors.Wrap(types.ErrUnknownDenom, denom)
	}

	// Decode ExchangeRate
	exchangeRate := &types.OracleExchangeRate{}
	k.cdc.MustUnmarshal(byteData, exchangeRate)
	return exchangeRate.ExchangeRate, exchangeRate.LastUpdate, exchangeRate.LastUpdateTimestamp, nil
}
