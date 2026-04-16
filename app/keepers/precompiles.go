package keepers

import (
	"fmt"
	"maps"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	transferkeeper "github.com/cosmos/ibc-go/v10/modules/apps/transfer/keeper"
	clientkeeper "github.com/cosmos/ibc-go/v10/modules/core/02-client/keeper"
	connectionkeeper "github.com/cosmos/ibc-go/v10/modules/core/03-connection/keeper"
	channelkeeper "github.com/cosmos/ibc-go/v10/modules/core/04-channel/keeper"

	"cosmossdk.io/core/address"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	bankprecompile "github.com/cosmos/evm/precompiles/bank"
	"github.com/cosmos/evm/precompiles/bech32"
	distprecompile "github.com/cosmos/evm/precompiles/distribution"
	govprecompile "github.com/cosmos/evm/precompiles/gov"
	ics20precompile "github.com/cosmos/evm/precompiles/ics20"
	"github.com/cosmos/evm/precompiles/p256"
	slashingprecompile "github.com/cosmos/evm/precompiles/slashing"
	stakingprecompile "github.com/cosmos/evm/precompiles/staking"
	erc20Keeper "github.com/cosmos/evm/x/erc20/keeper"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"

	"github.com/kiichain/kiichain/v7/precompiles/oracle"
	oraclekeeper "github.com/kiichain/kiichain/v7/x/oracle/keeper"
)

// Optionals define some optional params that can be applied to _some_ precompiles.
// Extend this struct, add a sane default to defaultOptionals, and an Option function to provide users with a non-breaking
// way to provide custom args to certain precompiles.
type Optionals struct {
	AddressCodec       address.Codec // used by gov/staking
	ValidatorAddrCodec address.Codec // used by slashing
	ConsensusAddrCodec address.Codec // used by slashing
}

// defaultOptionals returns the default coded optionals
func defaultOptionals() Optionals {
	return Optionals{
		AddressCodec:       evmAddressCodec{addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())},
		ValidatorAddrCodec: addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddrCodec: addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}
}

// evmAddressCodec wraps an account address codec and enforces that any decoded address is exactly
// 20 bytes (a valid EVM account).
//
// The stateful EVM precompiles accept an account address (e.g. the distribution withdraw address)
// and mirror the resulting bank transfer into the EVM StateDB via common.BytesToAddress, which is
// keyed by a 20-byte address. A longer account (e.g. a 32-byte bech32 account) would be silently
// truncated to its trailing 20 bytes during mirroring, causing the StateDB commit to mint a
// duplicate balance to that trailing-20-byte account and inflate native supply. Rejecting any
// non-20-byte address at decode time prevents such addresses from ever entering a mirrored flow.
type evmAddressCodec struct {
	address.Codec
}

// StringToBytes decodes the address with the wrapped codec and rejects any result that is not
// exactly 20 bytes, so only EVM-compatible accounts reach the balance-mirroring precompiles.
func (c evmAddressCodec) StringToBytes(text string) ([]byte, error) {
	bz, err := c.Codec.StringToBytes(text)
	if err != nil {
		return nil, err
	}
	if len(bz) != common.AddressLength {
		return nil, fmt.Errorf("invalid address %q: precompiles only accept 20-byte EVM accounts, got %d bytes", text, len(bz))
	}
	return bz, nil
}

// Option returns a funcion for the corresponding needed coded
type Option func(opts *Optionals)

// WithAddressCodec returns the function to access the with address codec
func WithAddressCodec(codec address.Codec) Option {
	return func(opts *Optionals) {
		opts.AddressCodec = codec
	}
}

// WithValidatorAddrCodec returns the function to access the with validator address codec
func WithValidatorAddrCodec(codec address.Codec) Option {
	return func(opts *Optionals) {
		opts.ValidatorAddrCodec = codec
	}
}

// WithConsensusAddrCodec returns the function to access the with consensus address codec
func WithConsensusAddrCodec(codec address.Codec) Option {
	return func(opts *Optionals) {
		opts.ConsensusAddrCodec = codec
	}
}

const bech32PrecompileBaseGas = 6_000

// NewAvailableStaticPrecompiles returns the list of all available static precompiled contracts from EVM.
//
// NOTE: this should only be used during initialization of the Keeper.
func NewAvailableStaticPrecompiles(
	stakingKeeper stakingkeeper.Keeper,
	distributionKeeper distributionkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20Keeper.Keeper,
	transferKeeper transferkeeper.Keeper,
	clientKeeper clientkeeper.Keeper,
	connectionKeeper connectionkeeper.Keeper,
	channelKeeper *channelkeeper.Keeper,
	evmKeeper *evmkeeper.Keeper,
	govKeeper govkeeper.Keeper,
	slashingKeeper slashingkeeper.Keeper,
	evidenceKeeper evidencekeeper.Keeper,
	wasmdKeeper wasmkeeper.Keeper,
	oracleKeeper oraclekeeper.Keeper,
	codec codec.Codec,
	opts ...Option,
) map[common.Address]vm.PrecompiledContract {
	// Set options
	options := defaultOptionals()
	for _, opt := range opts {
		opt(&options)
	}

	// Clone the mapping from the latest EVM fork.
	precompiles := maps.Clone(vm.PrecompiledContractsBerlin)

	// secp256r1 precompile as per EIP-7212
	p256Precompile := &p256.Precompile{}

	// Prepare the bech32 precompile
	bech32Precompile, err := bech32.NewPrecompile(bech32PrecompileBaseGas)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate bech32 precompile: %w", err))
	}

	// Prepare the staking precompile
	stakingPrecompile := stakingprecompile.NewPrecompile(
		stakingKeeper,
		stakingkeeper.NewMsgServerImpl(&stakingKeeper),
		stakingkeeper.NewQuerier(&stakingKeeper),
		bankKeeper,
		options.AddressCodec,
	)

	// Prepare the distribution precompile
	distributionPrecompile := distprecompile.NewPrecompile(
		distributionKeeper,
		distributionkeeper.NewMsgServerImpl(distributionKeeper),
		distributionkeeper.NewQuerier(distributionKeeper),
		stakingKeeper,
		bankKeeper,
		options.AddressCodec,
	)

	// Prepare the bank precompile
	bankPrecompile := bankprecompile.NewPrecompile(bankKeeper, erc20Keeper)

	ics20precompile := ics20precompile.NewPrecompile(
		bankKeeper,
		stakingKeeper,
		transferKeeper,
		channelKeeper,
		erc20Keeper,
	)

	// Prepare the gov precompile
	govPrecompile := govprecompile.NewPrecompile(
		govkeeper.NewMsgServerImpl(&govKeeper),
		govkeeper.NewQueryServer(&govKeeper),
		bankKeeper,
		codec,
		options.AddressCodec,
	)
	// Prepare the slashing precompile
	slashingPrecompile := slashingprecompile.NewPrecompile(
		slashingKeeper,
		slashingkeeper.NewMsgServerImpl(slashingKeeper),
		bankKeeper,
		options.ValidatorAddrCodec,
		options.ConsensusAddrCodec,
	)

	// Prepare the oracle precompile
	oraclePrecompile := oracle.NewPrecompile(oracleKeeper)

	// Stateless precompiles
	precompiles[bech32Precompile.Address()] = bech32Precompile
	precompiles[p256Precompile.Address()] = p256Precompile

	// Stateful precompiles
	precompiles[stakingPrecompile.Address()] = stakingPrecompile
	precompiles[distributionPrecompile.Address()] = distributionPrecompile
	precompiles[ics20precompile.Address()] = ics20precompile
	precompiles[bankPrecompile.Address()] = bankPrecompile
	precompiles[govPrecompile.Address()] = govPrecompile
	precompiles[slashingPrecompile.Address()] = slashingPrecompile
	precompiles[oraclePrecompile.Address()] = oraclePrecompile

	// Return the precompiles
	return precompiles
}
