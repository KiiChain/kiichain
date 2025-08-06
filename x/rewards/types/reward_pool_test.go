package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v4/x/rewards/types"
)

func TestRewardPoolValidateGenesis(t *testing.T) {
	rp := types.InitialRewardPool()
	require.Nil(t, rp.ValidateGenesis())

	rp2 := types.RewardPool{CommunityPool: sdk.DecCoins{{Denom: "tkii", Amount: math.LegacyNewDec(-1)}}}
	require.NotNil(t, rp2.ValidateGenesis())
}
