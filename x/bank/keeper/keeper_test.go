package keeper

import (
    "testing"

    sdk "cosmossdk.io/math"
    "github.com/stretchr/testify/require"
)

func TestBasicCoinMath(t *testing.T) {
    amt1 := sdk.NewInt(100)
    amt2 := sdk.NewInt(50)
    total := amt1.Add(amt2)

    require.Equal(t, sdk.NewInt(150), total, "Total harus 150")
}

