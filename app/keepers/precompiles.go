package keepers

import (
	"fmt"
	"maps"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

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
	transferkeeper "github.com/cosmos/evm/x/ibc/transfer/keeper"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"

	"github.com/kiichain/kiichain/v3/precompiles/ibc"
	"github.com/kiichain/kiichain/v3/precompiles/oracle"
	"github.com/kiichain/kiichain/v3/precompiles/wasmd"
	oraclekeeper "github.com/kiichain/kiichain/v3/x/oracle/keeper"
)

// Optionals define some optional params that can be applied to _some_ precompiles.
// Extend this struct, add a sane default to defaultOptionals, and an Option function to provide users with a non-breaking
// way to provide custom args to certain precompiles.
type Optionals struct {
	AddressCodec       address.Codec // used by gov/staking
	ValidatorAddrCodec address.Codec // used by slashing
	ConsensusAddrCodec address.Codec // used by slashing
}

func defaultOptionals() Optionals {
	return Optionals{
		AddressCodec:       addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddrCodec: addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddrCodec: addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}
}

type Option func(opts *Optionals)

func WithAddressCodec(codec address.Codec) Option {
	return func(opts *Optionals) {
		opts.AddressCodec = codec
	}
}

func WithValidatorAddrCodec(codec address.Codec) Option {
	return func(opts *Optionals) {
		opts.ValidatorAddrCodec = codec
	}
}

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
	stakingPrecompile, err := stakingprecompile.NewPrecompile(stakingKeeper, options.AddressCodec)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate staking precompile: %w", err))
	}

	// Prepare the distribution precompile
	distributionPrecompile, err := distprecompile.NewPrecompile(
		distributionKeeper,
		stakingKeeper,
		evmKeeper,
		options.AddressCodec,
	)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate distribution precompile: %w", err))
	}

	// Prepare the ibc precompile
	ibcTransferPrecompile, err := ics20precompile.NewPrecompile(
		bankKeeper,
		stakingKeeper,
		transferKeeper,
		channelKeeper,
		evmKeeper,
	)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate ICS20 precompile: %w", err))
	}

	// Prepare the bank precompile
	bankPrecompile, err := bankprecompile.NewPrecompile(bankKeeper, erc20Keeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate bank precompile: %w", err))
	}

	// Prepare the gov precompile
	govPrecompile, err := govprecompile.NewPrecompile(govKeeper, codec, options.AddressCodec)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate gov precompile: %w", err))
	}

	// Prepare the slashing precompile
	slashingPrecompile, err := slashingprecompile.NewPrecompile(slashingKeeper, options.ValidatorAddrCodec, options.ConsensusAddrCodec)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate slashing precompile: %w", err))
	}

	// Prepare the wasmd precompile
	wasmdPrecompile, err := wasmd.NewPrecompile(wasmdKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate wasmd precompile: %w", err))
	}

	// Prepare the ibc precompile
	ibcPrecompile, err := ibc.NewPrecompile(transferKeeper, clientKeeper, connectionKeeper, *channelKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate ibc precompile: %w", err))
	}

	// Prepare the oracle precompile
	oraclePrecompile, err := oracle.NewPrecompile(oracleKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate oracle precompile: %w", err))
	}

	// Stateless precompiles
	precompiles[bech32Precompile.Address()] = bech32Precompile
	precompiles[p256Precompile.Address()] = p256Precompile

	// Stateful precompiles
	precompiles[stakingPrecompile.Address()] = stakingPrecompile
	precompiles[distributionPrecompile.Address()] = distributionPrecompile
	precompiles[ibcTransferPrecompile.Address()] = ibcTransferPrecompile
	precompiles[bankPrecompile.Address()] = bankPrecompile
	precompiles[govPrecompile.Address()] = govPrecompile
	precompiles[slashingPrecompile.Address()] = slashingPrecompile
	precompiles[wasmdPrecompile.Address()] = wasmdPrecompile
	precompiles[ibcPrecompile.Address()] = ibcPrecompile
	precompiles[oraclePrecompile.Address()] = oraclePrecompile

	// Return the precompiles
	return precompiles
}
