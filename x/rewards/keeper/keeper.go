package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

type (
	Keeper struct {
		cdc codec.BinaryCodec

		storeService  store.KVStoreService
		bankKeeper    types.BankKeeper
		stakingKeeper types.StakingKeeper

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority        string
		feeCollectorName string // name of the FeeCollector ModuleAccount

		Schema     collections.Schema
		Params     collections.Item[types.Params]
		RewardPool collections.Item[types.RewardPool]
	}
)

// NewKeeper returns a new instance of the x/rewards keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
	authority, feeCollectorName string,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc: cdc,

		storeService:  storeService,
		bankKeeper:    bankKeeper,
		stakingKeeper: stakingKeeper,

		authority:        authority,
		feeCollectorName: feeCollectorName,

		Params:     collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		RewardPool: collections.NewItem(sb, types.RewardPoolKey, "reward_pool", codec.CollValue[types.RewardPool](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the x/rewards module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// StoreService returns the module KV store service (used by migrations).
func (k Keeper) StoreService() store.KVStoreService {
	return k.storeService
}

// Logger returns a logger for the x/rewards module
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ModuleAddress is the account that holds spendable reward funds.
// Fund it with a bank send; there is no fund-pool message.
func (k Keeper) ModuleAddress() sdk.AccAddress {
	return authtypes.NewModuleAddress(types.ModuleName)
}
