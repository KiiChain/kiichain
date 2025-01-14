package simulation_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	epochsimulation "github.com/kiichain/kiichain3/x/epoch/simulation"

	"github.com/stretchr/testify/require"
)

func TestFindAccount(t *testing.T) {
	// Setup
	var accs []simtypes.Account
	accs = append(accs, simtypes.Account{
		Address: sdk.AccAddress([]byte("kii1qzdrwc3806zfdl98608nqnsvhg8hn854jh6xyk")),
	})
	accs = append(accs, simtypes.Account{
		Address: sdk.AccAddress([]byte("kii1jdppe6fnj2q7hjsepty5crxtrryzhuqsxyeey6")),
	})

	// Test with account present
	addr1 := sdk.AccAddress([]byte("kii1qzdrwc3806zfdl98608nqnsvhg8hn854jh6xyk")).String()
	account, found := epochsimulation.FindAccount(accs, addr1)
	require.True(t, found)
	require.Equal(t, sdk.AccAddress([]byte("kii1qzdrwc3806zfdl98608nqnsvhg8hn854jh6xyk")), account.Address)

	// Test with account not present
	addr3 := sdk.AccAddress([]byte("address3")).String()
	account, found = epochsimulation.FindAccount(accs, addr3)
	require.False(t, found)
	require.Equal(t, simtypes.Account{}, account)

	// Test with invalid account address
	require.Panics(t, func() { epochsimulation.FindAccount(accs, "invalid") }, "The function did not panic with an invalid account address")
}
