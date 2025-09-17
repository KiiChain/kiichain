package kiichain

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/viper"

	clienthelpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	evmtypes "github.com/cosmos/evm/x/vm/types"

	"github.com/kiichain/kiichain/v4/app/params"
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
	params.LocalChainID: {
		Denom:         params.BaseDenom,
		ExtendedDenom: params.BaseDenom,
		DisplayDenom:  params.DisplayDenom,
		Decimals:      params.BaseDenomUnit,
	},
	params.DefaultChainID: {
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
		log.Printf("Chain ID %d not found in ChainsCoinInfo, using default", chainID)
		coinInfo = ChainsCoinInfo[params.LocalChainID]
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

var KiichainID uint64 = params.DefaultChainID // default Chain ID

// init initializes the KiichainID variable by reading the chain ID from the
// genesis file or app.toml file in the node's home directory.
// If the genesis file exists, it reads the Cosmos chain ID from there and parses it
// using the Evmos-style chain ID format; otherwise, it checks the app.toml file for the EVM chain ID.
// If neither file exists or the chain ID is not found, it defaults to the Kiichain Chain ID (262144).
func init() {
	nodeHome, err := clienthelpers.GetNodeHomeDirectory(".kiichain")
	if err != nil {
		panic(err)
	}

	// check if the genesis file exists and read the chain ID from it
	genesisFilePath := filepath.Join(nodeHome, "config", "genesis.json")
	if _, err = os.Stat(genesisFilePath); err == nil {
		// File exists, read the genesis file to get the chain ID
		reader, err := os.Open(genesisFilePath)
		if err == nil {
			defer reader.Close()

			chainID, err := genutiltypes.ParseChainIDFromGenesis(reader)
			if err == nil && chainID != "" {
				// Parse using Evmos-style chain ID format
				evmChainID, err := ParseChainID(chainID)
				if err == nil {
					KiichainID = evmChainID
					return
				}
				// If parsing fails, continue to check app.toml
			}
		}
	}
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	// If genesis file does not exist or chain ID is not found, check app.toml
	// to get the EVM chain ID
	appTomlPath := filepath.Join(nodeHome, "config", "app.toml")
	if _, err = os.Stat(appTomlPath); err == nil {
		// File exists
		v := viper.New()
		v.SetConfigFile(appTomlPath)
		v.SetConfigType("toml")

		if err = v.ReadInConfig(); err == nil {
			evmChainIDKey := "evm.evm-chain-id"
			if v.IsSet(evmChainIDKey) {
				evmChainID := v.GetUint64(evmChainIDKey)
				KiichainID = evmChainID
			}
		}
	}
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
}

var (
	regexChainID         = `[a-z]{1,}`
	regexEIP155Separator = `_{1}`
	regexEIP155          = `[1-9][0-9]*`
	regexEpochSeparator  = `-{1}`
	regexEpoch           = `[1-9][0-9]*`
	evmosChainID         = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)%s(%s)$`,
		regexChainID,
		regexEIP155Separator,
		regexEIP155,
		regexEpochSeparator,
		regexEpoch))
)

// ParseChainID parses a string chain identifier's EIP155 number to an Ethereum-compatible
// chain-id in uint64 format. The function returns an error if the chain-id has an invalid format
func ParseChainID(chainID string) (uint64, error) {
	chainID = strings.TrimSpace(chainID)
	if len(chainID) > 48 {
		return 0, fmt.Errorf("chain-id '%s' cannot exceed 48 chars", chainID)
	}

	matches := evmosChainID.FindStringSubmatch(chainID)
	if matches == nil || len(matches) != 4 || matches[1] == "" {
		return 0, fmt.Errorf("chain-id '%s' does not match Evmos format: %s_%s-%s", chainID, regexChainID, regexEIP155, regexEpoch)
	}

	// verify that the EIP155 part (matches[2]) is a base 10 integer
	chainIDInt, err := strconv.ParseUint(matches[2], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("EIP155 identifier '%s' must be base-10 integer format", matches[2])
	}

	return chainIDInt, nil
}
