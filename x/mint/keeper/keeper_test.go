package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	epochTypes "github.com/kiichain/kiichain3/x/epoch/types"
	"github.com/kiichain/kiichain3/x/mint/keeper"
	"github.com/kiichain/kiichain3/x/mint/types"

	"github.com/stretchr/testify/require"
)

type MockAccountKeeper struct {
	ModuleAddress       sdk.AccAddress
	ModuleAccount       authtypes.ModuleAccountI
	moduleNameToAddress map[string]string
}

func (m MockAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	if addrStr, ok := m.moduleNameToAddress[name]; ok {
		addr, _ := sdk.AccAddressFromBech32(addrStr)
		return addr
	}
	return nil
}

func (m MockAccountKeeper) SetModuleAccount(ctx sdk.Context, account authtypes.ModuleAccountI) {}

func (m MockAccountKeeper) GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI {
	return m.ModuleAccount
}

func (m MockAccountKeeper) SetModuleAddress(name, address string) {
	m.moduleNameToAddress[name] = address
}

type MockMintHooks struct {
	afterDistributeMintedCoinCalled bool
}

func (h *MockMintHooks) AfterDistributeMintedCoin(ctx sdk.Context, mintedCoin sdk.Coin) {
	h.afterDistributeMintedCoinCalled = true
}

func TestGetNextScheduledTokenRelease(t *testing.T) {
	t.Parallel()

	currentTime := time.Now().UTC()
	epoch := epochTypes.Epoch{
		CurrentEpochStartTime: currentTime,
		CurrentEpochHeight:    100,
	}
	currentMinter := types.DefaultInitialMinter()

	tokenReleaseSchedule := []types.ScheduledTokenRelease{
		{
			StartDate:          currentTime.AddDate(0, 0, 30).Format(types.TokenReleaseDateFormat),
			EndDate:            currentTime.AddDate(0, 2, 0).Format(types.TokenReleaseDateFormat),
			TokenReleaseAmount: 200,
		},
		{
			StartDate:          currentTime.AddDate(1, 0, 0).Format(types.TokenReleaseDateFormat),
			EndDate:            currentTime.AddDate(2, 0, 0).Format(types.TokenReleaseDateFormat),
			TokenReleaseAmount: 300,
		},
		{
			StartDate:          currentTime.AddDate(0, 0, 1).Format(types.TokenReleaseDateFormat),
			EndDate:            currentTime.AddDate(0, 0, 15).Format(types.TokenReleaseDateFormat),
			TokenReleaseAmount: 100,
		},
	}

	t.Run("Get the next scheduled token release", func(t *testing.T) {
		// No next scheduled token release intially
		epoch.CurrentEpochStartTime = currentTime.AddDate(0, 0, 0)
		nextScheduledRelease, err := keeper.GetNextScheduledTokenRelease(epoch, tokenReleaseSchedule, currentMinter)
		require.NoError(t, err)
		require.Nil(t, nextScheduledRelease)

		epoch.CurrentEpochStartTime = currentTime.AddDate(0, 0, 1)
		nextScheduledRelease, err = keeper.GetNextScheduledTokenRelease(epoch, tokenReleaseSchedule, currentMinter)
		require.NoError(t, err)
		require.NotNil(t, nextScheduledRelease)
		require.Equal(t, uint64(100), nextScheduledRelease.TokenReleaseAmount)
	})

	t.Run("No next scheduled token release, assume we are on the second period", func(t *testing.T) {
		// No next scheduled token release initially
		epoch.CurrentEpochStartTime = currentTime.AddDate(0, 0, 0)
		nextScheduledRelease, err := keeper.GetNextScheduledTokenRelease(epoch, tokenReleaseSchedule, currentMinter)
		require.NoError(t, err)
		require.Nil(t, nextScheduledRelease)

		secondMinter := types.NewMinter(
			currentTime.AddDate(0, 0, 30).Format(types.TokenReleaseDateFormat),
			currentTime.AddDate(0, 2, 0).Format(types.TokenReleaseDateFormat),
			"ukii",
			200,
		)
		epoch.CurrentEpochStartTime = currentTime.AddDate(0, 5, 0)
		nextScheduledRelease, err = keeper.GetNextScheduledTokenRelease(epoch, tokenReleaseSchedule, secondMinter)
		require.NoError(t, err)
		require.Nil(t, nextScheduledRelease)
	})

	t.Run("test case where we skip the start date due to outage for two days", func(t *testing.T) {
		// No next scheduled token release initially
		epoch.CurrentEpochStartTime = currentTime.AddDate(0, 0, 0)
		nextScheduledRelease, err := keeper.GetNextScheduledTokenRelease(epoch, tokenReleaseSchedule, currentMinter)
		require.NoError(t, err)
		require.Nil(t, nextScheduledRelease)

		// First mint was +1 but the chain recovered on +3
		epoch.CurrentEpochStartTime = currentTime.AddDate(0, 0, 3)
		nextScheduledRelease, err = keeper.GetNextScheduledTokenRelease(epoch, tokenReleaseSchedule, currentMinter)
		require.NoError(t, err)
		require.Equal(t, uint64(100), nextScheduledRelease.GetTokenReleaseAmount())
		require.Equal(t, currentTime.AddDate(0, 0, 1).Format(types.TokenReleaseDateFormat), nextScheduledRelease.GetStartDate())
	})

	t.Run("Error on invalid format", func(t *testing.T) {
		// No next scheduled token release initially
		tokenReleaseSchedule := []types.ScheduledTokenRelease{
			{
				StartDate:          "Bad Start Date",
				EndDate:            currentTime.AddDate(0, 2, 0).Format(types.TokenReleaseDateFormat),
				TokenReleaseAmount: 200,
			},
		}
		epoch.CurrentEpochStartTime = currentTime.AddDate(0, 0, 0)

		// Check for error
		_, err := keeper.GetNextScheduledTokenRelease(epoch, tokenReleaseSchedule, currentMinter)
		require.ErrorContains(t, err, "invalid scheduled release date")
	})

	t.Run("Error on bad minter", func(t *testing.T) {
		// Generate a bad minter
		badMinter := types.NewMinter(
			currentTime.AddDate(0, 0, 30).Format(types.TokenReleaseDateFormat),
			currentTime.AddDate(0, 2, 0).Format("BAD START DATE FORMAT"),
			"ukii",
			200,
		)

		// Check for error
		epoch.CurrentEpochStartTime = currentTime.AddDate(0, 5, 0)
		_, err := keeper.GetNextScheduledTokenRelease(epoch, tokenReleaseSchedule, badMinter)
		require.ErrorContains(t, err, "invalid end date for current minter")
	})
}

func TestGetOrUpdateLatestMinter(t *testing.T) {
	t.Parallel()
	app, ctx := createTestApp(false)
	appKeeper := app.MintKeeper
	currentTime := time.Now()
	epoch := epochTypes.Epoch{
		CurrentEpochStartTime: currentTime,
	}

	t.Run("No ongoing release", func(t *testing.T) {
		currentMinter, err := appKeeper.GetOrUpdateLatestMinter(ctx, epoch)
		require.NoError(t, err)
		require.False(t, currentMinter.OngoingRelease())
	})

	t.Run("No ongoing release, but there's a scheduled release", func(t *testing.T) {
		appKeeper.SetMinter(ctx, types.NewMinter(
			currentTime.Format(types.TokenReleaseDateFormat),
			currentTime.AddDate(1, 0, 0).Format(types.TokenReleaseDateFormat),
			"ukii",
			1000,
		))
		epoch.CurrentEpochStartTime = currentTime
		currentMinter, err := appKeeper.GetOrUpdateLatestMinter(ctx, epoch)
		require.NoError(t, err)
		require.True(t, currentMinter.OngoingRelease())
		require.Equal(t, currentTime.Format(types.TokenReleaseDateFormat), currentMinter.StartDate)
		appKeeper.SetMinter(ctx, types.DefaultInitialMinter())
	})

	t.Run("Ongoing release same day", func(t *testing.T) {
		params := appKeeper.GetParams(ctx)
		params.TokenReleaseSchedule = []types.ScheduledTokenRelease{
			{
				StartDate:          currentTime.AddDate(0, 0, 0).Format(types.TokenReleaseDateFormat),
				EndDate:            currentTime.AddDate(0, 0, 0).Format(types.TokenReleaseDateFormat),
				TokenReleaseAmount: 1000,
			},
		}
		appKeeper.SetParams(ctx, params)

		minter := types.Minter{
			StartDate:           currentTime.Format(types.TokenReleaseDateFormat),
			EndDate:             currentTime.Format(types.TokenReleaseDateFormat),
			Denom:               "ukii",
			TotalMintAmount:     100,
			RemainingMintAmount: 0,
			LastMintAmount:      100,
			LastMintDate:        "2023-04-01",
			LastMintHeight:      0,
		}
		appKeeper.SetMinter(ctx, minter)

		epoch.CurrentEpochStartTime = currentTime
		currentMinter, err := appKeeper.GetOrUpdateLatestMinter(ctx, epoch)
		require.NoError(t, err)
		releaseAmountToday, err := currentMinter.GetReleaseAmountToday(currentTime)
		require.NoError(t, err)
		amount := releaseAmountToday.IsZero()
		require.Zero(t, currentMinter.GetRemainingMintAmount())
		require.True(t, amount)
		require.False(t, currentMinter.OngoingRelease())
		require.Equal(t, currentTime.Format(types.TokenReleaseDateFormat), currentMinter.StartDate)
		appKeeper.SetMinter(ctx, types.DefaultInitialMinter())
	})

	t.Run("TokenReleaseSchedule not sorted", func(t *testing.T) {
		params := appKeeper.GetParams(ctx)
		params.TokenReleaseSchedule = []types.ScheduledTokenRelease{
			{
				StartDate:          currentTime.AddDate(0, 20, 0).Format(types.TokenReleaseDateFormat),
				EndDate:            currentTime.AddDate(0, 45, 0).Format(types.TokenReleaseDateFormat),
				TokenReleaseAmount: 2000,
			},
			{
				StartDate:          currentTime.Format(types.TokenReleaseDateFormat),
				EndDate:            currentTime.AddDate(0, 15, 0).Format(types.TokenReleaseDateFormat),
				TokenReleaseAmount: 1000,
			},
		}
		appKeeper.SetParams(ctx, params)

		epoch.CurrentEpochStartTime = currentTime
		currentMinter, err := appKeeper.GetOrUpdateLatestMinter(ctx, epoch)
		require.NoError(t, err)
		require.True(t, currentMinter.OngoingRelease())
		require.Equal(t, currentTime.Format(types.TokenReleaseDateFormat), currentMinter.StartDate)
	})
}

func TestBaseCases(t *testing.T) {
	t.Parallel()
	app, ctx := createTestApp(false)
	appKeeper := app.MintKeeper

	t.Run("invalid module name", func(t *testing.T) {
		mockAccountKeeper := MockAccountKeeper{}

		require.Panics(t, func() {
			keeper.NewKeeper(
				appKeeper.GetCdc(),
				appKeeper.GetStoreKey(),
				appKeeper.GetParamSpace(),
				nil,
				mockAccountKeeper,
				nil,
				nil,
				"invalid module",
			)
		})
	})

	t.Run("set hooks", func(t *testing.T) {
		newHook := &MockMintHooks{}
		appKeeper.SetHooks(newHook)

		require.PanicsWithValue(t, "cannot set mint hooks twice", func() {
			appKeeper.SetHooks(newHook)
		})
	})

	t.Run("nil minter", func(t *testing.T) {
		nilApp, nilCtx := createTestApp(false)

		store := nilCtx.KVStore(nilApp.MintKeeper.GetStoreKey())
		store.Delete(types.MinterKey)
		require.PanicsWithValue(t, "stored minter should not have been nil", func() {
			nilApp.MintKeeper.GetMinter(nilCtx)
		})
	})

	t.Run("staking keeper calls", func(t *testing.T) {
		require.False(t, appKeeper.StakingTokenSupply(ctx).IsNil())
		require.False(t, appKeeper.BondedRatio(ctx).IsNil())
	})

	t.Run("mint keeper calls", func(t *testing.T) {
		require.NotNil(t, appKeeper.GetStoreKey())
		require.NotNil(t, appKeeper.GetCdc())
		require.NotNil(t, appKeeper.GetParamSpace())
		require.NotPanics(t, func() {
			appKeeper.SetParamSpace(appKeeper.GetParamSpace())
		})
	})

	t.Run("staking keeper calls", func(t *testing.T) {
		require.Nil(t, appKeeper.MintCoins(ctx, sdk.NewCoins()))
	})

}
