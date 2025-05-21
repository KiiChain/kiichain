package ibc_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	testkeyring "github.com/cosmos/evm/testutil/integration/os/keyring"
	"github.com/cosmos/evm/x/vm/statedb"
	kiichainApp "github.com/kiichain/kiichain/v1/app"
	ibcprecompile "github.com/kiichain/kiichain/v1/precompiles/ibc"
)

// These integration tests were modified to work with the KiichainApp
// Sources:
// * Transfer tests: https://github.com/cosmos/ibc-go/blob/v7.3.2/modules/apps/29-fee/transfer_test.go#L13
// * ICA tests: https://github.com/cosmos/ibc-go/blob/v7.3.2/modules/apps/29-fee/ica_test.go#L94
var (
	// transfer + IBC fee test variables
	defaultRecvFee    = sdk.Coins{sdk.Coin{Denom: "stake", Amount: math.NewInt(100)}}
	defaultAckFee     = sdk.Coins{sdk.Coin{Denom: "stake", Amount: math.NewInt(200)}}
	defaultTimeoutFee = sdk.Coins{sdk.Coin{Denom: "stake", Amount: math.NewInt(300)}}
)

type IBCPrecompileTestSuite struct {
	suite.Suite
	coordinator *ibctesting.Coordinator
	keyring     testkeyring.Keyring

	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
	path   *ibctesting.Path

	// Precompile
	Precompile *ibcprecompile.Precompile
}

func TestIBCPrecompileTestSuite(t *testing.T) {
	ibcPrecompileSuite := &IBCPrecompileTestSuite{}
	suite.Run(t, ibcPrecompileSuite)
}

func (s *IBCPrecompileTestSuite) SetupTest() {
	// Set the DefaultBondDenom as default
	sdk.DefaultBondDenom = "stake"
	ibctesting.DefaultTestingAppInit = KiichainAppIniterTempDir

	// Start a new keyring
	keyring := testkeyring.New(2)
	s.keyring = keyring

	// Setup coordinator and chains
	s.coordinator = ibctesting.NewCoordinator(s.T(), 2)
	s.chainA = s.coordinator.GetChain(ibctesting.GetChainID(1))
	s.chainB = s.coordinator.GetChain(ibctesting.GetChainID(2))

	// Check if chain is KiichainApp
	chain, ok := s.coordinator.Chains[ibctesting.GetChainID(1)]
	s.Require().True(ok, "chain not found")
	_, ok = chain.App.(*kiichainApp.KiichainApp)
	s.Require().True(ok, "expected App to be KiichainApp")

	// Create path / channel betweed A and B
	path := ibctesting.NewPath(s.chainA, s.chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.MockFeePort
	path.EndpointB.ChannelConfig.PortID = ibctesting.MockFeePort
	s.path = path

	// Setup ibc precompile on chain A
	pc, err := ibcprecompile.NewPrecompile(
		chain.App.(*kiichainApp.KiichainApp).TransferKeeper, chain.App.GetIBCKeeper().ClientKeeper,
		chain.App.GetIBCKeeper().ConnectionKeeper, chain.App.GetIBCKeeper().ChannelKeeper)
	s.Require().NoError(err)
	s.Precompile = pc
}

// GetStateDB returns the state database for the precompile from a given chain
func (s *IBCPrecompileTestSuite) GetStateDB(chain *ibctesting.TestChain) *statedb.StateDB {
	ctx := chain.GetContext()
	// Get the header hash
	headerHash := ctx.HeaderHash()

	// Return the statedb
	return statedb.New(
		ctx,
		chain.App.(*kiichainApp.KiichainApp).EVMKeeper,
		statedb.NewEmptyTxConfig(common.BytesToHash(headerHash)),
	)
}
