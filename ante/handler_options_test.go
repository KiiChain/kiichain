//go:build test

package ante_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/runtime"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	cosmosevmantetypes "github.com/cosmos/evm/ante/types"

	kiiante "github.com/kiichain/kiichain/v7/ante"
	"github.com/kiichain/kiichain/v7/app/helpers"
)

func TestHandlerOptionsValidateRequiresCircuitKeeper(t *testing.T) {
	kiiApp := helpers.Setup(t)
	wasmConfig := wasmtypes.DefaultNodeConfig()
	opts := kiiante.HandlerOptions{
		Cdc:                    kiiApp.AppCodec(),
		AccountKeeper:          &kiiApp.AccountKeeper,
		BankKeeper:             kiiApp.BankKeeper,
		ExtensionOptionChecker: cosmosevmantetypes.HasDynamicFeeExtensionOption,
		EvmKeeper:              kiiApp.EVMKeeper,
		FeeAbstractionKeeper:   kiiApp.FeeAbstractionKeeper,
		FeegrantKeeper:         kiiApp.FeeGrantKeeper,
		CircuitKeeper:          &kiiApp.CircuitKeeper,
		IBCKeeper:              kiiApp.IBCKeeper,
		FeeMarketKeeper:        kiiApp.FeeMarketKeeper,
		SignModeHandler:        kiiApp.GetTxConfig().SignModeHandler(),
		SigGasConsumer:         kiiante.SigVerificationGasConsumer,
		DynamicFeeChecker:      true,
		StakingKeeper:          kiiApp.StakingKeeper,
		TXCounterStoreService:  runtime.NewKVStoreService(kiiApp.GetKey(wasmtypes.StoreKey)),
		WasmConfig:             &wasmConfig,
		OracleKeeper:           &kiiApp.OracleKeeper,
		PendingTxListener:      func(common.Hash) {},
	}
	require.NoError(t, opts.Validate())

	opts.CircuitKeeper = nil
	err := opts.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "circuit keeper is required for AnteHandler")
}
