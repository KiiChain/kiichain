package keeper_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/kiichain/kiichain/v7/x/rewards/keeper"
	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

type mockBankKeeper struct {
	sendErr error
}

func (m mockBankKeeper) SendCoinsFromModuleToModule(context.Context, string, string, sdk.Coins) error {
	return m.sendErr
}

func (m mockBankKeeper) SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error {
	return nil
}

func (m mockBankKeeper) GetBalance(context.Context, sdk.AccAddress, string) sdk.Coin {
	return sdk.Coin{}
}

type mockStakingKeeper struct {
	ratio math.LegacyDec
	err   error
}

func (m mockStakingKeeper) BondedRatio(context.Context) (math.LegacyDec, error) {
	return m.ratio, m.err
}

func setupRewardsKeeper(
	t *testing.T,
	bank types.BankKeeper,
	staking types.StakingKeeper,
) (sdk.Context, keeper.Keeper) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		bank,
		staking,
		authtypes.NewModuleAddress("gov").String(),
		authtypes.FeeCollectorName,
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{Time: time.Now().UTC()}, false, log.NewNopLogger())
	return ctx, k
}

func TestBeginBlockerErrorPaths(t *testing.T) {
	t.Run("params missing returns error", func(t *testing.T) {
		ctx, k := setupRewardsKeeper(t, mockBankKeeper{}, mockStakingKeeper{})
		err := k.BeginBlocker(ctx)
		require.Error(t, err)
	})

	t.Run("reward pool missing returns error", func(t *testing.T) {
		ctx, k := setupRewardsKeeper(t, mockBankKeeper{}, mockStakingKeeper{})
		params := types.DefaultParams()
		params.SupplyBase = math.NewInt(1_000_000)
		require.NoError(t, k.Params.Set(ctx, params))

		err := k.BeginBlocker(ctx)
		require.Error(t, err)
	})

	t.Run("bonded ratio error skips without failing block", func(t *testing.T) {
		ctx, k := setupRewardsKeeper(t,
			mockBankKeeper{},
			mockStakingKeeper{err: errors.New("staking unavailable")},
		)

		params := types.DefaultParams()
		params.SupplyBase = math.NewInt(1_000_000)
		require.NoError(t, k.Params.Set(ctx, params))

		now := ctx.BlockTime()
		require.NoError(t, k.RewardPool.Set(ctx, types.RewardPool{
			CommunityPool:   sdk.NewDecCoins(sdk.NewDecCoin(params.TokenDenom, math.NewInt(1_000_000))),
			LastReleaseTime: now.Add(-time.Hour),
		}))

		require.NoError(t, k.BeginBlocker(ctx))

		pool, err := k.RewardPool.Get(ctx)
		require.NoError(t, err)
		require.True(t, pool.TotalReleased.IsZero() || pool.TotalReleased.IsNil() || pool.TotalReleased.Amount.IsZero())
		// Clock advances so a later successful block does not dump accrued Δt
		require.True(t, pool.LastReleaseTime.Equal(now))
	})

	t.Run("bank send failure skips without failing block", func(t *testing.T) {
		ctx, k := setupRewardsKeeper(t,
			mockBankKeeper{sendErr: errors.New("insufficient funds")},
			mockStakingKeeper{ratio: math.LegacyNewDecWithPrec(30, 2)},
		)

		params := types.DefaultParams()
		params.SupplyBase = math.NewInt(1_000_000_000_000)
		require.NoError(t, k.Params.Set(ctx, params))

		now := ctx.BlockTime()
		require.NoError(t, k.RewardPool.Set(ctx, types.RewardPool{
			CommunityPool:   sdk.NewDecCoins(sdk.NewDecCoin(params.TokenDenom, math.NewInt(1_000_000))),
			LastReleaseTime: now.Add(-time.Hour),
		}))

		require.NoError(t, k.BeginBlocker(ctx))

		pool, err := k.RewardPool.Get(ctx)
		require.NoError(t, err)
		require.True(t, pool.LastReleaseTime.Equal(now))
		require.True(t, pool.TotalReleased.IsZero() || pool.TotalReleased.IsNil() || pool.TotalReleased.Amount.IsZero())
	})
}

func TestWriteRewardMetrics(t *testing.T) {
	ctx, k := setupRewardsKeeper(t, mockBankKeeper{}, mockStakingKeeper{})
	// Happy path: normal amounts convert cleanly
	k.WriteRewardMetrics(ctx, sdk.NewCoin("akii", math.NewInt(100)), sdk.NewCoin("akii", math.NewInt(200)))
}
