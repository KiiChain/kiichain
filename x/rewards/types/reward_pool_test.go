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
			name: "valid single coin",
			pool: types.RewardPool{
				CommunityPool: sdk.DecCoins{
					{Denom: "akii", Amount: math.LegacyNewDec(100)},
				},
			},
			expectErr: false,
		},
		{
			name: "negative amount",
			pool: types.RewardPool{
				CommunityPool: sdk.DecCoins{
					{Denom: "tkii", Amount: math.LegacyNewDec(-1)},
				},
			},
			expectErr: true,
		},
		{
			name: "invalid denom format (starts with digit)",
			pool: types.RewardPool{
				CommunityPool: sdk.DecCoins{
					{Denom: "1invalid", Amount: math.LegacyNewDec(1)},
				},
			},
			expectErr: true,
		},
		{
			name: "duplicate denoms",
			pool: types.RewardPool{
				CommunityPool: sdk.DecCoins{
					{Denom: "akii", Amount: math.LegacyNewDec(1)},
					{Denom: "akii", Amount: math.LegacyNewDec(2)},
				},
			},
			expectErr: true,
		},
		{
			name: "non-canonical ordering",
			pool: types.RewardPool{
				CommunityPool: sdk.DecCoins{
					{Denom: "tkii", Amount: math.LegacyNewDec(1)},
					{Denom: "akii", Amount: math.LegacyNewDec(2)},
				},
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
