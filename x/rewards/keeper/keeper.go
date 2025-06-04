package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v1/x/rewards/types"
)

type (
	Keeper struct {
		cdc codec.BinaryCodec

		accountKeeper       types.AccountKeeper
		bankKeeper          types.BankKeeper
		communityPoolKeeper types.CommunityPoolKeeper

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		Schema         collections.Schema
		Params         collections.Item[types.Params]
		RewardPool     collections.Item[types.RewardPool]
		RewardReleaser collections.Item[types.RewardReleaser]
	}
)

// NewKeeper returns a new instance of the x/tokenfactory keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	communityPoolKeeper types.CommunityPoolKeeper,
	authority string,
) Keeper {

	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc: cdc,

		accountKeeper:       accountKeeper,
		bankKeeper:          bankKeeper,
		communityPoolKeeper: communityPoolKeeper,

		authority:      authority,
		Params:         collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		RewardPool:     collections.NewItem(sb, types.RewardPoolKey, "reward_pool", codec.CollValue[types.RewardPool](cdc)),
		RewardReleaser: collections.NewItem(sb, types.RewardPoolKey, "reward_releaser", codec.CollValue[types.RewardReleaser](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the x/mint module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a logger for the x/tokenfactory module
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// FundCommunityPool allows an account to directly fund the community fund pool.
// The amount is first added to the rewards module account and then directly
// added to the pool. An error is returned if the amount cannot be sent to the
// module account.
func (k Keeper) FundCommunityPool(ctx context.Context, amount sdk.Coin, sender sdk.AccAddress) error {
	coins := sdk.Coins{amount}
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, coins); err != nil {
		return err
	}

	rewardPool, err := k.RewardPool.Get(ctx)
	if err != nil {
		return err
	}

	rewardPool.CommunityPool = rewardPool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(coins...)...)
	return k.RewardPool.Set(ctx, rewardPool)
}
