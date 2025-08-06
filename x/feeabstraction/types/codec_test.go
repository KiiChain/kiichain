package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
)

// TestRegisterInterfaces test the register interfaces
func TestRegisterInterfaces(t *testing.T) {
	// Initialize an empty registry
	registry := codectypes.NewInterfaceRegistry()
	registry.RegisterInterface(sdk.MsgInterfaceProtoName, (*sdk.Msg)(nil))

	// Run the register interfaces
	types.RegisterInterfaces(registry)

	// Check the interface registration result
	interfaces := registry.ListImplementations(sdk.MsgInterfaceProtoName)

	// Check the response
	require.ElementsMatch(t, interfaces, []string{
		"/kiichain.feeabstraction.v1beta1.MsgUpdateParams",
		"/kiichain.feeabstraction.v1beta1.MsgUpdateFeeTokens",
	})
}
