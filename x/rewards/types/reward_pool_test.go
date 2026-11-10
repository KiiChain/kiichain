package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

func TestRewardPoolValidateGenesis(t *testing.T) {
	testCases := []struct {
		name      string
		pool      types.RewardPool
		expectErr bool
	}{
		{
			name:      "initial empty pool",
			pool:      types.InitialRewardPool(),
			expectErr: false,
		},
		{
			name: "valid total released",
			pool: types.RewardPool{
				TotalReleased: sdk.NewCoin("akii", math.NewInt(100)),
			},
			expectErr: false,
		},
		{
			name: "invalid total released coin",
			pool: types.RewardPool{
				TotalReleased: sdk.Coin{Denom: "1bad", Amount: math.NewInt(1)},
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.pool.ValidateGenesis()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
