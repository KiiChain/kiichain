package kiichain

import (
	"log"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	evmtypes "github.com/cosmos/evm/x/vm/types"

	"github.com/kiichain/kiichain/v3/app/params"
)

// EVMOptionsFn defines a function type for setting app options specifically for
// the Cosmos EVM app. The function should receive the chainID and return an error if
// any.
type EVMOptionsFn func(uint64) error

// NoOpEVMOptions is a no-op function that can be used when the app does not
// need any specific configuration
func NoOpEVMOptions(_ uint64) error {
	return nil
}

var sealed = false

// ChainsCoinInfo is a map of the chain id and its corresponding EvmCoinInfo
// that allows initializing the app with different coin info based on the
// chain id
var ChainsCoinInfo = map[uint64]evmtypes.EvmCoinInfo{
	params.TestnetChainID: {
		Denom:         params.BaseDenom,
		ExtendedDenom: params.BaseDenom,
		DisplayDenom:  params.DisplayDenom,
		Decimals:      params.BaseDenomUnit,
	},
	0: {
		Denom:         params.BaseDenom,
		ExtendedDenom: params.BaseDenom,
		DisplayDenom:  params.DisplayDenom,
		Decimals:      params.BaseDenomUnit,
	},
}

// EVMAppOptions allows to setup the global configuration
// for the chain.
func EVMAppOptions(chainID uint64) error {
	// Check if the configuration is sealed
	if sealed {
		return nil
	}

	coinInfo, found := ChainsCoinInfo[chainID]
	if !found {
		// If not found, set as default
		log.Println("Chain ID not found in ChainsCoinInfo, using default")
		coinInfo = ChainsCoinInfo[params.TestnetChainID]
	}

	// set the denom info for the chain
	if err := setBaseDenom(coinInfo); err != nil {
		return err
	}

	ethCfg := evmtypes.DefaultChainConfig(chainID)

	err := evmtypes.NewEVMConfigurator().
		WithChainConfig(ethCfg).
		// NOTE: we're using the 18 decimals default for the example chain
		WithEVMCoinInfo(coinInfo).
		Configure()
	if err != nil {
		return err
	}

	sealed = true
	return nil
}

// setBaseDenom registers the display denom and base denom and sets the
// base denom for the chain.
func setBaseDenom(ci evmtypes.EvmCoinInfo) error {
	if err := sdk.RegisterDenom(ci.DisplayDenom, math.LegacyOneDec()); err != nil {
		return err
	}

	// sdk.RegisterDenom will automatically overwrite the base denom when the
	// new setBaseDenom() are lower than the current base denom's units.
	return sdk.RegisterDenom(ci.Denom, math.LegacyNewDecWithPrec(1, int64(ci.Decimals)))
}
