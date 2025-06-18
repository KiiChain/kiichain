package e2e

import (
	"fmt"
	"os"

	tmrand "github.com/cometbft/cometbft/libs/rand"

	dbm "github.com/cosmos/cosmos-db"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/types"

	"cosmossdk.io/log"
	evidencetypes "cosmossdk.io/x/evidence/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distribtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramsproptypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	erc20types "github.com/cosmos/evm/x/erc20/types"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	kiichain "github.com/kiichain/kiichain/v2/app"
	kiiparams "github.com/kiichain/kiichain/v2/app/params"
	rewardstypes "github.com/kiichain/kiichain/v2/x/rewards/types"
	tokenfactorytypes "github.com/kiichain/kiichain/v2/x/tokenfactory/types"
)

const (
	keyringPassphrase = "testpassphrase"
	keyringAppName    = "testnet"
)

var (
	encodingConfig kiiparams.EncodingConfig
	cdc            codec.Codec
	txConfig       client.TxConfig
)

func init() {
	encodingConfig = kiiparams.MakeEncodingConfig()
	banktypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	authtypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	authvesting.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	stakingtypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	evidencetypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	cryptocodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	govv1types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	govv1beta1types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	paramsproptypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	paramsproptypes.RegisterLegacyAminoCodec(encodingConfig.Amino)

	upgradetypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	distribtypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ratelimittypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	rewardstypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	tokenfactorytypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	// EVM register interfaces
	evmtypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	erc20types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	feemarkettypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	cdc = encodingConfig.Marshaler
	txConfig = encodingConfig.TxConfig
}

type chain struct {
	dataDir    string
	id         string
	validators []*validator
	accounts   []*account //nolint:unused
	// initial accounts in genesis
	genesisAccounts        []*account
	genesisVestingAccounts map[string]sdk.AccAddress
}

func newChain() (*chain, error) {
	tmpDir, err := os.MkdirTemp("", "kiichain-e2e-testnet-")
	if err != nil {
		return nil, err
	}

	return &chain{
		id:      fmt.Sprintf("%s-%d", kiiparams.LocalChainID, tmrand.Int()),
		dataDir: tmpDir,
	}, nil
}

func (c *chain) configDir() string {
	return fmt.Sprintf("%s/%s", c.dataDir, c.id)
}

func (c *chain) createAndInitValidators(count int) error {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}

	tempApplication := kiichain.NewKiichainApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		tmpDir,
		kiichain.EmptyAppOptions{},
		kiichain.EmptyWasmOptions,
		kiichain.NoOpEVMOptions,
	)
	defer func() {
		if err := tempApplication.Close(); err != nil {
			panic(err)
		}
	}()

	genesisState := tempApplication.ModuleBasics.DefaultGenesis(encodingConfig.Marshaler)

	for i := 0; i < count; i++ {
		node := c.createValidator(i)

		// generate genesis files
		if err := node.init(genesisState); err != nil {
			return err
		}

		c.validators = append(c.validators, node)

		// create keys
		if err := node.createKey("val"); err != nil {
			return err
		}
		if err := node.createNodeKey(); err != nil {
			return err
		}
		if err := node.createConsensusKey(); err != nil {
			return err
		}
	}

	return nil
}

func (c *chain) createAndInitValidatorsWithMnemonics(count int, mnemonics []string) error { //nolint:unused // this is called during e2e tests
	tempApplication := kiichain.NewKiichainApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		kiichain.DefaultNodeHome,
		kiichain.EmptyAppOptions{},
		kiichain.EmptyWasmOptions,
		kiichain.NoOpEVMOptions,
	)
	defer func() {
		if err := tempApplication.Close(); err != nil {
			panic(err)
		}
	}()

	genesisState := tempApplication.ModuleBasics.DefaultGenesis(encodingConfig.Marshaler)

	for i := 0; i < count; i++ {
		// create node
		node := c.createValidator(i)

		// generate genesis files
		if err := node.init(genesisState); err != nil {
			return err
		}

		c.validators = append(c.validators, node)

		// create keys
		if err := node.createKeyFromMnemonic("val", mnemonics[i]); err != nil {
			return err
		}
		if err := node.createNodeKey(); err != nil {
			return err
		}
		if err := node.createConsensusKey(); err != nil {
			return err
		}
	}

	return nil
}

func (c *chain) createValidator(index int) *validator {
	return &validator{
		chain:   c,
		index:   index,
		moniker: fmt.Sprintf("%s-kiichain-%d", c.id, index),
	}
}
