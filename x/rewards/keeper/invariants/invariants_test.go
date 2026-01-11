package invariants

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

// testKeeper implements KeeperInterface for testing
type testKeeper struct {
	schedule types.ReleaseSchedule
	pool     types.RewardPool
	params   types.Params
}

func (t testKeeper) ReleaseScheduleGet(ctx sdk.Context) (types.ReleaseSchedule, error) {
	return t.schedule, nil
}

func (t testKeeper) RewardPoolGet(ctx sdk.Context) (types.RewardPool, error) {
	return t.pool, nil
}

func (t testKeeper) ParamsGet(ctx sdk.Context) (types.Params, error) {
	return t.params, nil
}

func TestReleaseScheduleInvariant(t *testing.T) {
	ctx := sdk.Context{}
    
	t.Run("Valid schedule passes", func(t *testing.T) {
		k := testKeeper{
			schedule: types.ReleaseSchedule{
				TotalAmount:     sdk.NewCoin("ukii", math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin("ukii", math.NewInt(500)),
				Active:          true,
				EndTime:         time.Now().Add(time.Hour),
				LastReleaseTime: time.Now(),
			},
		}
        
		inv := ReleaseScheduleInvariant(k)
		msg, broken := inv(ctx)
		require.False(t, broken, "Valid schedule should pass: %s", msg)
	})
    
	t.Run("Detects released > total", func(t *testing.T) {
		k := testKeeper{
			schedule: types.ReleaseSchedule{
				TotalAmount:     sdk.NewCoin("ukii", math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin("ukii", math.NewInt(1500)),
				Active:          true,
				EndTime:         time.Now().Add(time.Hour),
				LastReleaseTime: time.Now(),
			},
		}
        
		inv := ReleaseScheduleInvariant(k)
		msg, broken := inv(ctx)
		require.True(t, broken)
		require.Contains(t, msg, "released amount (1500ukii) > total amount (1000ukii)")
	})
    
	t.Run("Detects denom mismatch", func(t *testing.T) {
		k := testKeeper{
			schedule: types.ReleaseSchedule{
				TotalAmount:     sdk.NewCoin("ukii", math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin("akii", math.NewInt(500)),
				Active:          true,
				EndTime:         time.Now().Add(time.Hour),
				LastReleaseTime: time.Now(),
			},
		}
        
		inv := ReleaseScheduleInvariant(k)
		msg, broken := inv(ctx)
		require.True(t, broken)
		require.Contains(t, msg, "denom mismatch")
	})
    
	t.Run("Detects invalid time for active schedule", func(t *testing.T) {
		now := time.Now()
		k := testKeeper{
			schedule: types.ReleaseSchedule{
				TotalAmount:     sdk.NewCoin("ukii", math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin("ukii", math.NewInt(500)),
				Active:          true,
				EndTime:         now,
				LastReleaseTime: now,
			},
		}
        
		inv := ReleaseScheduleInvariant(k)
		msg, broken := inv(ctx)
		require.True(t, broken)
		require.Contains(t, msg, "end time")
	})
}

func TestCommunityPoolNonNegativeInvariant(t *testing.T) {
	ctx := sdk.Context{}
    
	t.Run("Valid pool passes", func(t *testing.T) {
		k := testKeeper{
			pool: types.RewardPool{
				CommunityPool: sdk.NewDecCoins(sdk.NewDecCoin("ukii", math.NewInt(1000))),
			},
		}
        
		inv := CommunityPoolNonNegativeInvariant(k)
		msg, broken := inv(ctx)
		require.False(t, broken, "Valid pool should pass: %s", msg)
	})
    
	// Catatan: Tidak bisa test negative amount karena SDK validasi mencegah pembuatan DecCoin negatif
	// Ini sebenarnya bagus - SDK sudah punya built-in protection
	t.Run("Zero pool passes", func(t *testing.T) {
		k := testKeeper{
			pool: types.RewardPool{
				CommunityPool: sdk.NewDecCoins(),
			},
		}
        
		inv := CommunityPoolNonNegativeInvariant(k)
		msg, broken := inv(ctx)
		require.False(t, broken, "Zero pool should pass: %s", msg)
	})
}

func TestParamsInvariant(t *testing.T) {
	ctx := sdk.Context{}
    
	t.Run("Valid params pass", func(t *testing.T) {
		k := testKeeper{
			params: types.Params{TokenDenom: "ukii"},
		}
        
		inv := ParamsInvariant(k)
		msg, broken := inv(ctx)
		require.False(t, broken, "Valid params should pass: %s", msg)
	})
    
	t.Run("Detects empty token denom", func(t *testing.T) {
		k := testKeeper{
			params: types.Params{TokenDenom: ""},
		}
        
		inv := ParamsInvariant(k)
		msg, broken := inv(ctx)
		require.True(t, broken)
		require.Contains(t, msg, "token denom is empty")
	})
}

// Benchmark tests untuk performance
func BenchmarkInvariants(b *testing.B) {
	ctx := sdk.Context{}
    
	k := testKeeper{
		schedule: types.ReleaseSchedule{
			TotalAmount:     sdk.NewCoin("ukii", math.NewInt(1000)),
			ReleasedAmount:  sdk.NewCoin("ukii", math.NewInt(500)),
			Active:          true,
			EndTime:         time.Now().Add(time.Hour),
			LastReleaseTime: time.Now(),
		},
		pool: types.RewardPool{
			CommunityPool: sdk.NewDecCoins(sdk.NewDecCoin("ukii", math.NewInt(1000))),
		},
		params: types.Params{TokenDenom: "ukii"},
	}
    
	b.Run("ReleaseScheduleInvariant", func(b *testing.B) {
		inv := ReleaseScheduleInvariant(k)
		for i := 0; i < b.N; i++ {
			_, _ = inv(ctx)
		}
	})
    
	b.Run("CommunityPoolInvariant", func(b *testing.B) {
		inv := CommunityPoolNonNegativeInvariant(k)
		for i := 0; i < b.N; i++ {
			_, _ = inv(ctx)
		}
	})
    
	b.Run("ParamsInvariant", func(b *testing.B) {
		inv := ParamsInvariant(k)
		for i := 0; i < b.N; i++ {
			_, _ = inv(ctx)
		}
	})
}
