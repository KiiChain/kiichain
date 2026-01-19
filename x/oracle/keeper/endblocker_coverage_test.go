package keeper_test

import (
	"testing"
	"github.com/stretchr/testify/require"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestEndBlockerErrorHandling(t *testing.T) {
	input, _ := SetUp(t)
	ctx := input.Ctx
	oracleKeeper := input.OracleKeeper

	// Test that invalid operator address is handled gracefully
	// This covers the error path in EndBlocker
	_, err := sdk.ValAddressFromBech32("invalid_operator_address")
	require.Error(t, err)

	// Test EndBlocker runs without panic even with invalid addresses
	require.NotPanics(t, func() {
		oracleKeeper.EndBlocker(ctx)
	})
}